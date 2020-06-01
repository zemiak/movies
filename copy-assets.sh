#!/bin/sh

# copies various static assets from node modules to public folder

third=public/thirdparty

copy() {
    from="${1}"
    to="${third}/${2}"
    mkdir -p ${to}
    cp -r ${from}/* ${to}/
}

rm -rf ${third}

copy node_modules/bulma/css bulma
copy node_modules/@fortawesome/fontawesome-free/css fontawesome
copy node_modules/@fortawesome/fontawesome-free/webfonts webfonts
