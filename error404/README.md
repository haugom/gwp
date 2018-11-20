# Build custom 404 error page

Just for fun. Copied a 404 page and build a simple go app to serve the content.


## build

`docker build --no-cache -t test .`

## test

`docker run -it --rm -P test`

## tag & push

`docker tag test:latest haugom/error404:0.1`

`docker push !$`
