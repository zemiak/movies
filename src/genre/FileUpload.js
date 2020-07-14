import { Config } from "../config/Config.js";

export class FileUpload {
    constructor(img, file, urlPart, id) {
        this.img = img;
        this.file = file;
        this.urlPart = urlPart;
        this.id = id;
        this.config = new Config();

        console.log("FileUpload.constructor()");

        this.upload();
    }

    upload() {
        var data = new FormData();
        data.append('file', this.file);
        data.append('id', this.id);

        this.ctrl = document.querySelector("#thumbnailProgress");
        const xhr = new XMLHttpRequest();
        this.xhr = xhr;

        const self = this;
        this.xhr.upload.addEventListener("progress", function(e) {
                if (e.lengthComputable) {
                    const percentage = Math.round((e.loaded * 100) / e.total);
                    self.ctrl.value = percentage;
                    console.log("FileUpload.upload()/percentage " + percentage);
                }
            }, false);

        xhr.upload.addEventListener("load", function(e){
                self.ctrl.value = "100";
            }, false);

        const url = this.getBaseUri() + "/" + this.urlPart + "/thumbnail";

        console.log("FileUpload.upload()/" + url);

        xhr.open("POST", url);
        xhr.send();
    }

    getBaseUri() {
        var base = this.config.getBackendUrl();
        if (base.includes("#")) {
            base = base.split('#')[0];
        }

        return base;
    }
}
