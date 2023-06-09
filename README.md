# gshr

> Git static host repo.

Command line tool for generating stand-alone, static git html hosting. Produces a single output
directory for multiple repos, with...

* Root index.html that lists all input repos.
* Commit log page for each repo.
* Individual commit page summarizing commit including diff.
* File list page for each repo for the current HEAD ref.
* File detail/preview page for each file in current HEAD ref.
* Statically clone-able git dir for each repo.

---

See for yourself.

```bash
git clone https://github.com/vogtb/gshr
cd gshr
make dev-example-gshr-simple
```

Which basically runs this.

```bash
./target/gshr.bin -c=example-config-gshr-simple.toml -o=target/output
cd target/output
python3 -m http.server 80
```

See example TOML configs in root directory.

---

## Usage

```
Usage of gshr:
  -c Config file.
  -o Dir of output.
  -s Run in silent mode.
```

The toml file needs to be in the format:

* `site`: Site-level configuration.
  * `base_url`: Base url for the site. Eg: `"/"` or `"https://mysite.com/code/"`.
  * `name`: Site name displayed on the main index.html page that lists all repos.
* `repos` List of data for each repo.
  * `name`: Name of repo to be used in display, paths.
  * `description`: Description of repo used in html pages.
  * `url`: Absolute, relative, or remote. eg: `/home/repo`, `./repo`, `git://`, `http://`.
  * `alt_link`: Optional Link to where the repo lives. Eg: `github.com/vogtb/gshr`.

Example:

```toml
[site]
base_url = "http://localhost/"
name = "development site - should run on port 80"

[[repos]]
name = "gshr"
description = "git static host repo -- generates static html for repos"
url = "https://github.com/vogtb/gshr"
alt_link = "https://github.com/vogtb/gshr"
```

---

## Output

```text
{output_dir}/
  index.html
  {repo_name}/
    log.html
    commits/
      {hash}/commit.html
    files.html
    files/
      {full_file_name}/file.html
  {repo_name}.git
      {...raw git file data}
```

Example:

```text
output
├── favicon.ico
├── gshr
│   ├── commit
│   │   ├── 069606b3fcd2f96fc4349943326fb31f9d3c561f
│   │   │   └── index.html
│   │   │   ...
│   │   │   ...
│   │   └── fe47541cb62d6f513734089afda72ddefe672924
│   │       └── index.html
│   ├── files
│   │   ├── LICENSE
│   │   │   └── index.html
│   │   ├── Makefile
│   │   │   └── index.html
│   │   ├── README.md
│   │   │   └── index.html
│   │   │   ...
│   │   │   ...
│   │   ├── main.go
│   │   │   └── index.html
│   │   └── templates
│   │       ├── commit.html
│   │       │   └── index.html
│   │       │   ...
│   │       │   ...
│   │       └── partials.html
│   │           └── index.html
│   ├── files.html
│   └── log.html
├── gshr.css
├── gshr.git
│   ├── HEAD
│   ├── config
│   ├── index
│   ├── info
│   │   └── refs
│   ├── objects
│   │   ├── info
│   │   │   └── packs
│   │   └── pack
│   │       ├── pack-a6e75f15316a2d809290159b8bdc88303c8090cb.idx
│   │       └── pack-a6e75f15316a2d809290159b8bdc88303c8090cb.pack
│   └── refs
│       ├── heads
│       │   └── main
│       ├── remotes
│       │   └── origin
│       │       └── main
│       └── tags
└── index.html
```

---

# License

The MIT License (MIT)

Copyright (c) 2023 Ben Vogt

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.