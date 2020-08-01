import { html, render } from "lit-html";

export class RenderItunes {
    constructor(itunesCallback, itunesSaveCallback) {
        this.chooseItunesArtwork = this.chooseItunesArtwork.bind(this);
        this.artworkItemClick = this.artworkItemClick.bind(this);
        this.artworkListClose = this.artworkListClose.bind(this);
        this.artworkEmptyClose = this.artworkEmptyClose.bind(this);
        this.artworkDetailClose = this.artworkDetailClose.bind(this);
        this.artworkDetailSave = this.artworkDetailSave.bind(this);
        this.itunesCallback = itunesCallback;
        this.itunesSaveCallback = itunesSaveCallback;
    }

    chooseItunesArtwork(event) {
        this.itunesCallback(document.querySelector("#movieName").value);
    }

    artworkItemClick(event) {
        var anchor = event.target.closest("a");
        var id = anchor.getAttribute("data-id");
        var url = this.itunesData[id].artworkUrl;
        render(html`<center><figure class="image is-128x128"><img src="${url}"></figure></center>`, document.querySelector("#itunesDetail"));
        document.querySelector("#itunesDetailTitle").innerText = this.itunesData[id].trackName;
        document.querySelector("#itunesModalDetail").classList.toggle("is-active");
    }

    artworkListClose() {
        document.querySelector("#itunesModal").classList.toggle("is-active");
    }

    artworkEmptyClose() {
        document.querySelector("#itunesModalEmpty").classList.toggle("is-active");
    }

    artworkDetailClose() {
        document.querySelector("#itunesModalDetail").classList.toggle("is-active");
    }

    artworkDetailSave() {
        var image = document.querySelector("#itunesDetail>center>figure>img");
        this.itunesSaveCallback(image.src);
    }

    updateItunesThumbnails(data) {
        if (0 == data.length) {
            document.querySelector("#itunesTitle").innerText = document.querySelector("#movieName").value;
            document.querySelector("#itunesModalEmpty").classList.toggle("is-active");
            return;
        }

        this.itunesData = data;

        var items = [];
        var i = 0;
        data.forEach(element => {
            items.push(html`<a href="#" class="panel-block" style="height: 50px;" data-id="${i}" id="itunes_${i}">
            <span class="panel-icon">
              <figure class="image is-16x16"><img src="${element.artworkUrl}"></figure>
            </span>
            <div style="padding-left: 1em;">${element.trackName}</div>
          </a>`);
            i++;
        });
        render(html`<nav class="panel">${items}</nav>`, document.querySelector("#itunesList"));
        for (var i = 0; i < data.length; i++) {
            document.querySelector("#itunes_" + i).addEventListener("click", event => {this.artworkItemClick(event)})
        }
        document.querySelector("#itunesModal").classList.toggle("is-active");
    }

    view(readOnly) {
        if (readOnly) {
            return html``;
        }

        var button = html`<button type="button" class="button is-link is-light" @click="${this.chooseItunesArtwork}">iTunes Thumbnail</button>`;
        var listOfArtwork = html`<div class="modal" id="itunesModal">
            <div class="modal-background"></div>
            <div class="modal-card">
            <header class="modal-card-head">
                <p class="modal-card-title">iTunes Artwork</p>
                <button class="delete" aria-label="close" type="button" @click="${this.artworkListClose}"></button>
            </header>
            <section class="modal-card-body">
                <div id="itunesList"></div>
            </section>
            </div>
        </div>`;
        var emptyResult = html`<div class="modal" id="itunesModalEmpty">
            <div class="modal-background"></div>
            <div class="modal-card">
            <header class="modal-card-head">
                <p class="modal-card-title">iTunes Artwork</p>
                <button class="delete" aria-label="close" type="button" @click="${this.artworkEmptyClose}"></button>
            </header>
            <section class="modal-card-body">
                The iTunes service returned an empty result for <b><span id="itunesTitle"></span></b>.
            </section>
            </div>
        </div>`;
        var artworkDetail = html`<div class="modal" id="itunesModalDetail">
            <div class="modal-background"></div>
            <div class="modal-card">
            <header class="modal-card-head">
                <p class="modal-card-title" id="itunesDetailTitle"></p>
                <button class="delete" aria-label="close" type="button" @click="${this.artworkDetailClose}"></button>
            </header>
            <section class="modal-card-body">
                <div id="itunesDetail" style="height:230px;"></div>
            </section>
            <footer class="modal-card-foot">
                <button type="button" class="button is-success" @click="${this.artworkDetailSave}">Save</button>
                <button type="button" class="button" @click="${this.artworkDetailClose}">Cancel</button>
            </footer>
            </div>
        </div>`;

        return html`${button} ${listOfArtwork} ${emptyResult} ${artworkDetail}`;
    }

    showSuccessItunesThumbnail() {
        document.querySelector("#itunesModalDetail").classList.toggle("is-active");
        document.querySelector("#itunesModal").classList.toggle("is-active");
        document.querySelector("#successMessage").classList.remove("is-hidden");
        document.querySelector("#successMessage>div>p").innerText = "Thumbnail updated";

        let thumbnail = document.querySelector("#thumbnailDisplay");
        window.setTimeout(_ => {thumbnail.src = thumbnail.src + "&t=" + new Date().getTime();}, 500);
    }

    showErrorItunesThumbnail(err) {
        document.querySelector("#itunesModalDetail").classList.toggle("is-active");
        document.querySelector("#itunesModal").classList.toggle("is-active");
        document.querySelector("#errorMessage").classList.remove("is-hidden");
        document.querySelector("#errorMessage>div>p").innerText = err.ok === false
            ? (err.status + " " + err.statusText)
            : err;
    }
}
