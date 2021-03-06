import { html, render } from "lit-html";

export class RenderImdb {
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
    }

    chooseButtonClicked(event) {
        window._RenderImdb_detailClick = this.artworkItemClick;
        let currentName = document.querySelector("#movieName").value

        if (this.lastName != currentName) {
            this.lastName = currentName;
            this.fetchMetadataCallback(currentName);
        } else {
            document.querySelector("#imdbModal").classList.toggle("is-active");
        }
    }

    artworkItemClick(anchor) {
        var id = anchor.getAttribute("data-id");
        var url = this.imdbData[id].imageUrl;
        render(html`<center><figure class="image is-128x128"><img src="${url}"></figure></center>`, document.querySelector("#imdbDetail"));
        document.querySelector("#imdbDetailTitle").innerText = this.imdbData[id].description;
        document.querySelector("#imdbDetailTitle").setAttribute("data-id", id);
        document.querySelector("#imdbModalDetail").classList.toggle("is-active");
    }

    artworkListClose() {
        document.querySelector("#imdbModal").classList.toggle("is-active");
    }

    artworkEmptyClose() {
        document.querySelector("#imdbModalEmpty").classList.toggle("is-active");
    }

    artworkDetailClose() {
        document.querySelector("#imdbModalDetail").classList.toggle("is-active");
    }

    artworkDetailSave() {
        var anchor = document.querySelector("#imdbDetailTitle");
        var id = anchor.getAttribute("data-id");
        var data = this.imdbData[id];
        this.metadataSaveCallback(data);
    }

    updateImdbThumbnails(data) {
        if (0 == data.length) {
            document.querySelector("#imdbTitle").innerText = document.querySelector("#movieName").value;
            document.querySelector("#imdbModalEmpty").classList.toggle("is-active");
            return;
        }

        this.imdbData = data;

        var items = [];
        var i = 0;
        data.forEach(element => {
            items.push(html`<a href="javascript:void(0)" onclick="javascript:window._RenderImdb_detailClick(this)" class="panel-block" style="height: 50px;" data-id="${i}">
            <span class="panel-icon">
              <figure class="image is-16x16"><img src="${element.imageUrl}"></figure>
            </span>
            <div style="padding-left: 1em;">${element.description}</div>
          </a>`);
            i++;
        });
        render(html`<nav class="panel">${items}</nav>`, document.querySelector("#imdbList"));
        document.querySelector("#imdbModal").classList.toggle("is-active");
    }

    view(readOnly) {
        if (readOnly) {
            return html``;
        }

        var button = html`<button type="button" class="button is-link is-light" @click="${this.chooseButtonClicked}">IMDB</button>`;
        var listOfArtwork = html`<div class="modal" id="imdbModal">
            <div class="modal-background"></div>
            <div class="modal-card">
            <header class="modal-card-head">
                <p class="modal-card-title">IMDB Artwork</p>
                <button class="delete" aria-label="close" type="button" @click="${this.artworkListClose}"></button>
            </header>
            <section class="modal-card-body">
                <div id="imdbList"></div>
            </section>
            </div>
        </div>`;
        var emptyResult = html`<div class="modal" id="imdbModalEmpty">
            <div class="modal-background"></div>
            <div class="modal-card">
            <header class="modal-card-head">
                <p class="modal-card-title">IMDB Artwork</p>
                <button class="delete" aria-label="close" type="button" @click="${this.artworkEmptyClose}"></button>
            </header>
            <section class="modal-card-body">
                The IMDB service returned an empty result for <b><span id="imdbTitle"></span></b>.
            </section>
            </div>
        </div>`;
        var artworkDetail = html`<div class="modal" id="imdbModalDetail">
            <div class="modal-background"></div>
            <div class="modal-card">
            <header class="modal-card-head">
                <p class="modal-card-title" id="imdbDetailTitle" data-id="x"></p>
                <button class="delete" aria-label="close" type="button" @click="${this.artworkDetailClose}"></button>
            </header>
            <section class="modal-card-body">
                <div id="imdbDetail" style="height:230px;"></div>
            </section>
            <footer class="modal-card-foot">
                <button type="button" class="button is-success" @click="${this.artworkDetailSave}">Save</button>
                <button type="button" class="button" @click="${this.artworkDetailClose}">Cancel</button>
            </footer>
            </div>
        </div>`;

        return html`${button} ${listOfArtwork} ${emptyResult} ${artworkDetail}`;
    }

    showSuccessImdbThumbnail(response) {
        document.querySelector("#imdbModalDetail").classList.toggle("is-active");
        document.querySelector("#imdbModal").classList.toggle("is-active");
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

    showErrorImdbThumbnail(err) {
        document.querySelector("#imdbModalDetail").classList.toggle("is-active");
        document.querySelector("#imdbModal").classList.toggle("is-active");
        document.querySelector("#errorMessage").classList.remove("is-hidden");
        document.querySelector("#errorMessage>div>p").innerText = err.ok === false
            ? (err.status + " " + err.statusText)
            : err;
    }
}
