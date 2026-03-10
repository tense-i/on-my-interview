# Experiential Memory (Pitfalls & Playbooks)

## [2026-03-11] Go 测试文件名里带 `_windows` 会被当成 Windows 平台文件
**Trigger**: 在 `/Users/zh/project/githubProj/on-my-interview/server/internal/storage/mysql` 新增了名为 `usage_windows_test.go` 的测试文件。  
**Symptom**: `go test` 和 `go list` 都完全看不到这份测试文件，包里只显示旧测试；`go list -json` 里它出现在 `IgnoredGoFiles`。  
**Diagnosis**: Go 会把文件名中的 `_windows` 识别成 `GOOS=windows` 条件后缀，所以 `usage_windows_test.go` 只会在 Windows 环境参与编译。  
**Fix**: 改名为不带平台后缀的文件名，例如 `usage_window_test.go`。  
**Prevention**: 新建 Go 文件时避免把 `aix/android/darwin/freebsd/linux/windows` 等平台关键字放在文件名后缀位置，尤其是 `_windows.go` / `_linux.go` 这种模式。  

## [2026-03-11] OpenAI-compatible `response_format=json_object` 对 DeepSeek 不能只给松散字段名
**Trigger**: 在 `/Users/zh/project/githubProj/on-my-interview/server/internal/llm/openai/extractor.go` 中，只用一句 `Return JSON only with schema_version, company, post_type, sentiment...` 去约束 DeepSeek 的结构化输出。  
**Symptom**: 真实联调时，DeepSeek 返回语义正确但 schema 不兼容的 JSON，比如 `company` 是字符串、`sentiment` 是字符串、`questions` 里是 `question_number/question_text`，最终报错 `decode structured payload: json: cannot unmarshal string into Go struct field StructuredPost.company of type llm.Company`。  
**Diagnosis**: OpenAI-compatible 不等于“严格遵守你脑中的结构”；如果 prompt 没把嵌套对象和每个字段的 shape 写清楚，模型会按自己更自然的 JSON 组织方式返回。  
**Fix**: 在 system prompt 里显式写出完整 schema 示例，并明确要求 `company`、`sentiment`、`key_events`、`questions` 的对象结构；同时包含小写 `json`，避免 `response_format=json_object` 的提供方校验差异。  
**Prevention**: 对任何需要稳定反序列化的 LLM 输出，把 prompt 视作 API contract，测试里直接断言关键 schema 片段存在，不要只测“服务端返回了某个理想 JSON”。  

## [2026-03-11] 牛客搜索接口会返回空的 `contentData` 记录
**Trigger**: 使用 `/Users/zh/project/githubProj/on-my-interview/server/internal/crawler/nowcoder/client.go` 调牛客搜索接口抓 `面经`。  
**Symptom**: 搜索结果里偶尔会混入 `id/title/content/createTime/editTime` 全空的记录；如果不拦截，会把空帖子写进 `raw_posts`，出现空 `source_post_id`、空标题、空正文。  
**Diagnosis**: 牛客搜索返回的 `records` 并不保证每条都是完整帖子数据，存在占位或异常记录。  
**Fix**: 在 crawler 层对 `contentData.id/title/content` 做 `TrimSpace` 校验，任一为空就直接跳过。  
**Prevention**: 对外部搜索/列表 API，不要默认“列表项一定可入库”；入库前先做主键字段和最小正文完整性校验。  

## [2026-03-11] MySQL nullable varchar 不能直接扫描到 Go string
**Trigger**: 在 `/Users/zh/project/githubProj/on-my-interview/server/internal/storage/mysql/repository.go` 中，把允许 `NULL` 的 `company_name_raw` / `company_name_norm` 直接 `Scan` 到 Go `string` 字段。  
**Symptom**: 运行抓取任务或调用 `GET /api/v1/posts` 时失败，错误为 `converting NULL to string is unsupported`。  
**Diagnosis**: 如果 SQL 列允许 `NULL`，而仓储层直接把它扫描到 `string`，在命中已有记录或做左连接查询时会触发运行时错误。  
**Fix**: 先扫描到 `sql.NullString`，再把 `.String` 回填到结构体字段。  
**Prevention**: 对任何数据库里可空的字符串列，先检查 schema，再决定是否用 `sql.NullString`；尤其是左连接字段和业务上“可识别为空”的字段。  

## [2026-03-10] `go:embed` 不能引用父目录
**Trigger**: 在 `/Users/zh/project/githubProj/on-my-interview/server/internal/storage/mysql` 中尝试用 `//go:embed ../../../migrations/001_init.sql` 嵌入 migration 文件。  
**Symptom**: `go test` / `go build` 直接失败，报错 `pattern ../../../migrations/001_init.sql: invalid pattern syntax`。  
**Diagnosis**: 如果 `go:embed` pattern 含有 `..`，Go 编译器会拒绝；嵌入资源必须位于当前 package 目录或其子目录下。  
**Fix**: 把需要嵌入的资源复制或移动到 package 子目录，例如 `server/internal/storage/mysql/migrations/001_init.sql`，再用 `//go:embed migrations/001_init.sql`。  
**Prevention**: 设计内嵌 schema、模板、静态资源时，优先把资源文件放在消费它的 package 子目录，避免依赖跨目录引用。  

# 示例
## [2026-03-07] Go build 在当前工作区会因为 VCS stamping 失败
**Trigger**: 在 `/Users/zh/project/githubProj/stock-market-simulator/server` 里执行 `go build ./...`  
**Symptom**: 报错 `error obtaining VCS status: exit status 128`，并提示 `Use -buildvcs=false to disable VCS stamping.`  
**Diagnosis**: 直接运行 `cd server && GOTOOLCHAIN=go1.26.0 go build ./...`；如果立刻出现上述报错，说明当前工作区不能被 Go 正常识别为可取 VCS 元数据的仓库根。  
**Fix**: 使用 `cd server && GOTOOLCHAIN=go1.26.0 go build -buildvcs=false ./...` 完成构建验证。  
**Prevention**: 在这个工作区里把 `-buildvcs=false` 作为 backend 构建/验证命令的默认参数，除非后续仓库根结构被修正。  
