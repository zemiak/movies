#!/bin/sh

#    "mount:bulma": "mount node_modules/bulma/css --to /css/bulma",
#    "mount:fontawesome_css": "mount node_modules/@fortawesome/fontawesome/css --to /css/fontawesome",
#    "mount:fontawesome_webfonts": "mount node_modules/@fortawesome/fontawesome/webfonts --to /css/webfonts"

copy() {
    from="${1}"
    to="public/thirdparty/${2}"
    mkdir -p ${to}
    cp -r ${from}/* ${to}/
}

rm -rf public/thirdparty
copy node_modules/bulma/css bulma
copy node_modules/@fortawesome/fontawesome-free/css fontawesome
copy node_modules/@fortawesome/fontawesome-free/webfonts webfonts
