# OpenClaw Memory Integration

To enable OpenClaw to store its memory using the `yazi` storage engine, you can provide the following instructions to OpenClaw. OpenClaw will use the `yazi_ctl` CLI tool to persist and retrieve its memory data.

## Requirements

1. Make sure `yazi` server is running.
2. Build and install `yazi_ctl` so it's available in OpenClaw's PATH:
   ```bash
   go build -o yazi_ctl cmd/cli/yazi_ctl.go
   sudo mv yazi_ctl /usr/local/bin/
   ```

## MEMORY.md Instructions

Add the following section to your `MEMORY.md` or your system prompt for OpenClaw:

```markdown
# External Memory Storage Rules

You have access to an external memory storage engine called `yazi`. 
Whenever you need to store long-term context, conversation history, or update your memory files (like `MEMORY.md` or daily logs), you MUST use the `yazi_ctl memory` commands to persist them.

### Commands to use:

1. **Save memory to yazi:**
   When you create or update a memory file, immediately sync it to the yazi storage engine using:
   `yazi_ctl memory save <key> <file_path>`
   *Example: `yazi_ctl memory save memory_2026_03_27 memory/2026-03-27.md`*

2. **Load memory from yazi:**
   When you need to recall past context that is not in your current workspace, you can load it from yazi:
   `yazi_ctl memory load <key> <file_path>`
   *Example: `yazi_ctl memory load memory_2026_03_27 restored_memory.md`*
   You can also print the memory directly to stdout by omitting the file path:
   `yazi_ctl memory load <key>`

**Important:** Do not use `yazi_ctl set` or `yazi_ctl get` for memory files, as memory files are often large Markdown texts. Always use the `yazi_ctl memory save` and `yazi_ctl memory load` subcommands.
```
