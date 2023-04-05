# gshr

> Git static host repo.

Command line tool for generating stand-alone, static git html hosting. Produces a single output
directory for multiple repos, with html files for most preview-able text files, commit log, and
more.

---

## Usage

```
Usage of gshr:
  -clone string
    	Dir to clone into. Default is /tmp/${rand}
  -config string
    	Config file.
  -output string
    	Dir of output.
  -silent
    	Run in silent mode.
```

The toml file needs to be in the format:

* `base_url`: String for base url that this site will be served from. Eg: `"/"` or
  `"https://mysite.com/git/"`.  
* `site_name`: String overall site name. Displayed on the main index.html page that lists all
  repos.
* `repos` List of data for each repo.
  * `name`: String for rendering the name of this repo.
  * `description`: String for rendering the description.
  * `url`: String of the local absolute path, `git://`, `http://`, or `https://` url of the repo.
  * `published_git_url`: String of where the repo lives. Eg: `git@github.com:vogtb/gshr.git`

---

## Output

```text
{output_dir}
  index.html
  {repo_name}
    log.html
    commits
      {hash}/commit.html
    files.html
    files
      {full_file_name}/file.html
```

For example:

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
└── index.html
```
