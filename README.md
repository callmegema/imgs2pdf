convert images to pdf
# Usage
`go run main.go /your/images/folder/path/ trim recursive compress`

Example: `go run main.go /your/books/images/folder/path/ r2l true 50`

Example: `go run main.go /your/book/images/folder/path/ l2r false 100`

## trim
- r2l     Trim the wide page from right to left
- l2r     Trim the wide page from left to right
- false   No trimming

## recursive
- true    Make PDFs Recursively
- false   Make a single PDF

## compress
- 100     compress quality(%), 100==no compress

# Future works
## ocr
- jp
- eng

# Test
`go test .`
