# Registry Fake Pusher

[![MIT license](https://img.shields.io/github/license/mashape/apistatus.svg)](https://opensource.org/licenses/MIT)

## Usage
With this tool, if you want to overlay new image layers from one repository tag in Registry on an repository tag in another repository in Registry, you do not need to download all the layers into docker daemon.

> Environment variables in source image will be ignored.

## Supports

- registry V2 API;


## Steps 

1. Specify SourceRegistry, SourceRepository and SourceTag (need get the top ImageLayer of the image)

1. Specify TargetRegistry, TargetRepository and TargetTag (the image location wishing to overlay an ImageLayer)

1. Download source and target manifest according to specified image locationes

1. Construct the to-overlay-imageLayer

1. Overlay the to-overlay-imageLayer to target manifest, generating a new manifest 

1. Copy the content of the imageLayer from sourceRegistry to targetRegistry if needed

1. Push the new manifest.
