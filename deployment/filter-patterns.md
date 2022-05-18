# Filter Patterns

Filter patterns use simple glob-style wildcard patterns. A single asterisk `*` will match any character except `/`. A double asterisk will match all the descendants of path.

|                        | `/home` | `/home/*` | `/home/**` |
| ---------------------- | :-----: | :-------: | :--------: |
| `/home`                | **✔**   | **✘**     | **✘**      |
| `/home/alice`          | **✘**   | **✔**     | **✔**      |
| `/home/alice/projects` | **✘**   | **✘**     | **✔**      |

NOTE: Filter patterns ending in a wildcard *will not* match the parent path. You need to add both the parent path, and the child patterns.

```terraform
# To exclude /home and all its descendants:
EXCLUDED_EXPORTS = ["/home", "/home/**"]
```

| Special Terms | Meaning
| ------------- | -------
| `*`           | matches any sequence of characters except `/`
| `/**`         | matches zero or more directories
| `?`           | matches any single character except `/`
| `[class]`     | matches any single character except `/` against a class of characters
| `{alt1,...}`  | matches a sequence of characters if one of the comma-separated alternatives matches

### Character Classes

Character classes support the following:

| Class      | Meaning
| ---------- | -------
| `[abc]`    | matches any single character within the set
| `[a-z]`    | matches any single character in the range
| `[^class]` | matches any single character which does *not* match the class
| `[!class]` | same as `^`: negates the class

### Combining include and exclude patterns

Include and exclude patterns can be combined. For an export to be accepted (and re-exported), the export *must* match an include pattern, and *must not* match an exclude pattern.
