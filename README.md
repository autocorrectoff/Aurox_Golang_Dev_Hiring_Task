# SimpleSitemapGenerator
Sitemap (https://www.sitemaps.org) generator command line tool - **WIP**

## Usage

### Running the script with arguments - Example
```
go run main.go -url=https://sitemaps.org/ -parallel=3 -max-depth=5 -output-file=test.xml
```

### Or you can build into an executable and then run the program - Example
```
go build
./main -url=https://sitemaps.org/ -parallel=3 -max-depth=5 -output-file=test.xml
```

