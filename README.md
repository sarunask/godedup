# godedup
Finds duplicate files in specified path

Has 3 main GO routines, which are interconnected as follows:

walker -> makeSha1Sum -> compare

## Installation

```bash
go get github.com/sarunask/godedup
```

## Usage

```bash
godedup -search_path /home/
```

## Help

```bash
godedup --help
```