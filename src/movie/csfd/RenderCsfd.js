import { html, render } from "lit-html";
import { MetadataDetailDialog } from "../metadata/MetadataDetailDialog.js";
import { MetadataEmptyDialog } from "../metadata/MetadataEmptyDialog.js";
import { MetadataListDialog } from "../metadata/MetadataListDialog.js";
import { MetadataItem } from "../metadata/MetadataItem.js";

export class RenderCsfd {
    constructor(fetchMetadataCallback, metadataSaveCallback) {
        this.chooseButtonClicked = this.chooseButtonClicked.bind(this);
        this.artworkItemClick = this.artworkItemClick.bind(this);
        this.artworkListClose = this.artworkListClose.bind(this);
        this.artworkEmptyClose = this.artworkEmptyClose.bind(this);
        this.artworkDetailClose = this.artworkDetailClose.bind(this);
        this.artworkDetailSave = this.artworkDetailSave.bind(this);
        this.fetchMetadataCallback = fetchMetadataCallback;
        this.metadataSaveCallback = metadataSaveCallback;
        this.lastName = Math.random();
        this.title = "CSFD";
    }

    chooseButtonClicked(event) {
        window._RenderCsfd_detailClick = this.artworkItemClick;
        let currentName = document.querySelector("#movieName").value

        if (this.lastName != currentName) {
            this.lastName = currentName;
            this.fetchMetadataCallback(currentName);
        } else {
            document.querySelector("#csfdModal").classList.toggle("is-active");
        }
    }

    artworkItemClick(anchor) {
        var id = anchor.getAttribute("data-id");
        var url = this.csfdData[id].imageUrl;
        render(html`<center><figure class="image is-128x128"><img src="${url}"></figure></center>`, document.querySelector("#csfdDetail"));
        document.querySelector("#csfdDetailTitle").innerText = this.csfdData[id].description;
        document.querySelector("#csfdDetailTitle").setAttribute("data-id", id);
        document.querySelector("#csfdModalDetail").classList.toggle("is-active");
    }

    artworkListClose() {
        document.querySelector("#csfdModal").classList.toggle("is-active");
    }

    artworkEmptyClose() {
        document.querySelector("#csfdModalEmpty").classList.toggle("is-active");
    }

    artworkDetailClose() {
        document.querySelector("#csfdModalDetail").classList.toggle("is-active");
    }

    artworkDetailSave() {
        var anchor = document.querySelector("#csfdDetailTitle");
        var id = anchor.getAttribute("data-id");
        var data = this.csfdData[id];
        this.metadataSaveCallback(data);
    }

    updateCsfdThumbnails(data) {
        if (0 == data.length) {
            document.querySelector("#csfdTitle").innerText = document.querySelector("#movieName").value;
            document.querySelector("#csfdModalEmpty").classList.toggle("is-active");
            return;
        }

        this.csfdData = data;

        var items = [];
        var i = 0;
        data.forEach(element => {
            items.push(html`<metadata-item data-image-url="${element.imageUrl}"
                data-description="${element.description}"
                data-url="${element.url}"
                data-title="${this.title}"></metadata-item>`);
            i++;
        });
        render(html`<nav class="panel">${items}</nav>`, document.querySelector(this.title + "-list"));
        document.querySelector("#" + this.title + "-list-dialog").on();
    }

    view(readOnly) {
        if (readOnly) {
            return html``;
        }

        let button = html`<button type="button" class="button is-link is-light" @click="${this.chooseButtonClicked}">${this.title}</button>`;
        let listOfArtwork = html`<metadata-list-dialog data-title="${this.title}"></metadata-list-dialog>`;
        let emptyResult = html`<metadata-empty-dialog data-title="${this.title}"></metadata-empty-dialog>`;
        let artworkDetail = html`<metadata-detail-dialog data-title="${this.title}"></metadata-detail-dialog>`;

        return html`${button} ${listOfArtwork} ${emptyResult} ${artworkDetail}`;
    }

    showSuccessCsfdThumbnail(response) {
        document.querySelector("#csfdModalDetail").classList.toggle("is-active");
        document.querySelector("#csfdModal").classList.toggle("is-active");
        document.querySelector("#successMessage").classList.remove("is-hidden");
        document.querySelector("#successMessage>div>p").innerText = "Metadata updated";

        response.json().then(data => {
            document.querySelector("#movieDescription").value = data.description;
            if (data.year) {
                document.querySelector("#movieYear").value = data.year;
            }
        })

        let thumbnail = document.querySelector("#thumbnailDisplay");
        window.setTimeout(_ => {thumbnail.src = thumbnail.src + "&t=" + new Date().getTime();}, 500);
    }

    showErrorCsfdThumbnail(err) {
        document.querySelector("#csfdModalDetail").classList.toggle("is-active");
        document.querySelector("#csfdModal").classList.toggle("is-active");
        document.querySelector("#errorMessage").classList.remove("is-hidden");
        document.querySelector("#errorMessage>div>p").innerText = err.ok === false
            ? (err.status + " " + err.statusText)
            : err;
    }
}
