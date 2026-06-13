---
title: "Quick start"
description: "Run your first comick command."
weight: 30
---

Once `comick` is on your `PATH`:

```bash
comick --help       # see the command tree
comick version      # build info
```

This is a fresh scaffold, so the command tree is just `version` for now. Add
your first real command in `cli/`, build on the `comick` library package,
and document it here.

A good first command usually fetches one thing and prints it as JSON, so the
output pipes straight into `jq` and the rest of your tools.
