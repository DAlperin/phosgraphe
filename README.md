# Phosgraphe

A rudimentary self-hosted cloudinary alternative. So far only supports width and height transforms but more coming soon.

## Stack

1. Go (service)
2. Postgres (metadata storage)
3. S3 (image storage/serving)

## High level overview

1. I call the api to upload image, give it a name. For this example lets imagine we upload an image called `demo` in the namespace `demo-namespace`
2. To retrieve the image as uploaded call `http://phosgraph.local/image/demo-namespace/demo`, this can be used as a regular image link, i.e for `\<img src="..."">`
3. In order to fit the image elsewhere I want a version of it sized 100x100, so I would call `http://phosgraph.local/image/demo-namespace/demo/height_100,width_100`
   1. The first time this link is called it may have a second or so of delay as the image is generated to specification
   2. But on future calls the previously generated will be retrieved as stored within usually around `2ms`
   3. Transforms are largely transitive, so calling `http://phosgraph.local/image/demo-namespace/demo/width_100,height_100`, still returns the previously generated image even though the order of transforms switches (this is not guaranteed it is based on my very unscientific hashing algo that probably needs some more work before it's production ready)