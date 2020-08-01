import { html, render } from "lit-html";

export class RenderCsfd {
    constructor(csfdCallback, csfdSaveCallback) {
        this.chooseCsfdArtwork = this.chooseCsfdArtwork.bind(this);
        this.artworkItemClick = this.artworkItemClick.bind(this);
        this.artworkListClose = this.artworkListClose.bind(this);
        this.artworkEmptyClose = this.artworkEmptyClose.bind(this);
        this.artworkDetailClose = this.artworkDetailClose.bind(this);
        this.artworkDetailSave = this.artworkDetailSave.bind(this);
        this.csfdCallback = csfdCallback;
        this.csfdSaveCallback = csfdSaveCallback;
    }

    chooseCsfdArtwork(event) {
        this.csfdCallback(document.querySelector("#movieName").value);
    }

    artworkItemClick(event) {
        var anchor = event.target.closest("a");
        var id = anchor.getAttribute("data-id");
        var url = this.csfdData[id].artworkUrl;
        render(html`<center><figure class="image is-128x128"><img src="${url}"></figure></center>`, document.querySelector("#csfdDetail"));
        document.querySelector("#csfdDetailTitle").innerText = this.csfdData[id].trackName;
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
        var image = document.querySelector("#csfdDetail>center>figure>img");
        this.csfdSaveCallback(image.src);
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
            items.push(html`<a href="#" class="panel-block" style="height: 50px;" data-id="${i}" id="csfd_${i}">
            <span class="panel-icon">
              <figure class="image is-16x16"><img src="${element.artworkUrl}"></figure>
            </span>
            <div style="padding-left: 1em;">${element.trackName}</div>
          </a>`);
            i++;
        });
        render(html`<nav class="panel">${items}</nav>`, document.querySelector("#csfdList"));
        for (var i = 0; i < data.length; i++) {
            document.querySelector("#csfd_" + i).addEventListener("click", event => {this.artworkItemClick(event)})
        }
        document.querySelector("#csfdModal").classList.toggle("is-active");
    }

    view(readOnly) {
        if (readOnly) {
            return html``;
        }

        var button = html`<button type="button" class="button is-link is-light" @click="${this.chooseCsfdArtwork}">csfd Thumbnail</button>`;
        var listOfArtwork = html`<div class="modal" id="csfdModal">
            <div class="modal-background"></div>
            <div class="modal-card">
            <header class="modal-card-head">
                <p class="modal-card-title">csfd Artwork</p>
                <button class="delete" aria-label="close" type="button" @click="${this.artworkListClose}"></button>
            </header>
            <section class="modal-card-body">
                <div id="csfdList"></div>
            </section>
            </div>
        </div>`;
        var emptyResult = html`<div class="modal" id="csfdModalEmpty">
            <div class="modal-background"></div>
            <div class="modal-card">
            <header class="modal-card-head">
                <p class="modal-card-title">csfd Artwork</p>
                <button class="delete" aria-label="close" type="button" @click="${this.artworkEmptyClose}"></button>
            </header>
            <section class="modal-card-body">
                The csfd service returned an empty result for <b><span id="csfdTitle"></span></b>.
            </section>
            </div>
        </div>`;
        var artworkDetail = html`<div class="modal" id="csfdModalDetail">
            <div class="modal-background"></div>
            <div class="modal-card">
            <header class="modal-card-head">
                <p class="modal-card-title" id="csfdDetailTitle"></p>
                <button class="delete" aria-label="close" type="button" @click="${this.artworkDetailClose}"></button>
            </header>
            <section class="modal-card-body">
                <div id="csfdDetail" style="height:230px;"></div>
            </section>
            <footer class="modal-card-foot">
                <button type="button" class="button is-success" @click="${this.artworkDetailSave}">Save</button>
                <button type="button" class="button" @click="${this.artworkDetailClose}">Cancel</button>
            </footer>
            </div>
        </div>`;

        return html`${button} ${listOfArtwork} ${emptyResult} ${artworkDetail}`;
    }

    showSuccessCsfdThumbnail() {
        document.querySelector("#csfdModalDetail").classList.toggle("is-active");
        document.querySelector("#csfdModal").classList.toggle("is-active");
        document.querySelector("#successMessage").classList.remove("is-hidden");
        document.querySelector("#successMessage>div>p").innerText = "Thumbnail updated";

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
