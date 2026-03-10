Paragraph before the blockquote lists that is long enough to need rewrapping to fit within one hundred characters of column width.

> - The registry handles connection pooling and lifecycle management for all database backends used by the ingestion pipeline
> - Short one
> - When a plugin fails to initialize the system logs the error context with the full stack trace and falls back to the default handler configuration

> * Using star markers inside a blockquote where the text is definitely long enough to exceed one hundred characters and need rewrapping
> * Another star item

> 1. The scheduler processes each pending task in priority order and retries on transient failure before marking the task as permanently failed
> 2. Brief entry
> 3. Dependency resolution walks the full import graph before scheduling any compilation work to detect and report circular references early

> > - Doubly nested blockquote list item that should be rewrapped and needs to be long enough to actually exceed the one hundred character column limit

> A plain blockquote paragraph sitting between list blocks that is long enough to verify it also gets rewrapped properly at the column boundary.

> - First item with enough text to wrap across the hundred character boundary when you account for the prefix overhead
>
> - Second item separated by a blank blockquote line which means goldmark treats these as separate list items in the parsed AST
