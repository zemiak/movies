import { html, render } from "lit-html";
import { FileUpload } from "/_dist_/files/FileUpload.js";

export class RenderMovieDetail {
    constructor(successMessage, errorMessage) {
        this.successMessage = successMessage;
        this.errorMessage = errorMessage;
        this.genreChanged = this.genreChanged.bind(this);
        this.selectItunesThumbnailClicked = this.selectItunesThumbnailClicked.bind(this);
        this.selectItunesThumbnailSelectClicked = this.selectItunesThumbnailSelectClicked.bind(this);
        this.selectItunesThumbnailCloseClicked = this.selectItunesThumbnailCloseClicked.bind(this);
        this.selectItunesEmptyThumbnailCloseClicked = this.selectItunesEmptyThumbnailCloseClicked.bind(this);
        this.selectItunesDetailThumbnailCloseClicked = this.selectItunesDetailThumbnailCloseClicked.bind(this);
        this.selectItunesDetailThumbnailSaveClicked = this.selectItunesDetailThumbnailSaveClicked.bind(this);
        window._RenderMovieDetail = this;
    }

    view(entity, readOnly, isNew) {
        let items = [];

        items.push(this.message(this.successMessage, "is-primary", "successMessage"));
        items.push(this.message(this.errorMessage, "is-danger", "errorMessage"));

        if (! isNew) {
            items.push(this.id(readOnly, entity.id));
        }

        items.push(this.text("movieName", "Name", readOnly, entity.name));
        items.push(this.text("movieOriginalName", "Original Name", readOnly, entity.originalName));
        items.push(this.textArea("movieDescription", "Description", readOnly, entity.description));
        items.push(this.genre(readOnly, entity.genreId, entity.genres));
        items.push(this.combo("movieSerieId", "Serie", readOnly, entity.serieId, entity.series));
        items.push(this.combo("movieLanguageId", "Language", readOnly, entity.languageId, entity.languages));
        items.push(this.combo("movieOriginalLanguageId", "OriginalLanguage", readOnly, entity.originalLanguageId, entity.languages));
        items.push(this.combo("movieSubtitlesId", "Subtitles", readOnly, entity.subtitlesId, entity.languages));
        items.push(this.number("movieDisplayOrder", "Display Order", readOnly, entity.displayOrder));
        items.push(this.hidden("moviePictureFileName", entity.pictureFileName));
        items.push(this.hidden("movieCreated", entity.created));
        items.push(this.hidden("movieFileName", entity.fileName));
        items.push(this.hidden("movieUrl", entity.url));
        items.push(this.hidden("movieWebPage", entity.webPage));
        items.push(this.hidden("movieYear", entity.year));
        items.push(this.thumbnail(readOnly, entity.thumbnailUrl));
        items.push(this.selectItunesThumbnailModal(readOnly));

        this.data = entity;

        let title = this.title(isNew ? "New Movie" : "Movie");

        return html`${title}<form class="form-horizontal"><fieldset>${items}</fieldset></form><p>&nbsp;</p>`;
    }

    getFormData() {
        var item = {};

        this.addItem(item, "id", "movieId");
        this.addItem(item, "name", "movieName");
        this.addItem(item, "originalName", "movieOriginalName");
        this.addItem(item, "description", "movieDescription");
        this.addItem(item, "genreId", "movieGenreId");
        this.addItem(item, "serieId", "movieSerieId");
        this.addItem(item, "languageId", "movieLanguageId");
        this.addItem(item, "originalLanguageId", "movieOriginalLanguageId");
        this.addItem(item, "subtitlesId", "movieSubtitlesId");
        this.addItem(item, "displayOrder", "movieDisplayOrder");
        this.addItem(item, "pictureFileName", "moviePictureFileName");
        this.addItem(item, "created", "movieCreated");
        this.addItem(item, "fileName", "movieFileName");
        this.addItem(item, "url", "movieUrl");
        this.addItem(item, "webPage", "movieWebPage");
        this.addItem(item, "year", "movieYear");

        return item;
    }

    addItem(data, name, elementId) {
        var element = document.querySelector("#" + elementId);
        data[name] = element ? element.value : null;
    }

    message(text, className, idName) {
        return html`<article id="${idName}" class="message ${className} is-hidden">
        <div class="message-header">
            <p>${text}</p>
        </div>
      </article>`;
    }

    thumbnail(readOnly, imageUrl) {
        var upload = readOnly ? "" : html`<div class="file">
        <label id="thumbnail" class="file-label">
          <input class="file-input" type="file" name="resume" id="thumbnail" accept="image/jpeg">
          <span class="file-cta">
            <span class="file-icon">
              <i class="fas fa-upload"></i>
            </span>
            <span class="file-label">
              Choose a file…
            </span>
          </span>
          <span class="file-name">
      JPG images only
    </span>
        </label>
      </div>
      <div style="padding-top: 1em;"><progress id="thumbnailProgress" max="100" value="0" class="progress is-success is-hidden"></progress></div>`;
        return html`<figure class="image is-128x128">
        <img class="is-rounded" id="thumbnailDisplay" src="${imageUrl}">
      </figure>${upload}`;
    }

    combo(id, label, readOnly, itemId, items,) {
        var templates = [];
        for (const [key, value] of Object.entries(items)) {
            templates.push(html`<option value="${key}" ?selected=${key == itemId}>${value}</option>`);
        }

        return html`<div class="field"><label class="label">${label}</label><div class="select"><select id="${id}" ?disabled=${readOnly}>${templates}</select></div></div>`;
    }

    genre(readOnly, itemId, items,) {
        var templates = [];
        for (const [key, value] of Object.entries(items)) {
            templates.push(html`<option value="${key}" ?selected=${key == itemId}>${value}</option>`);
        }

        return html`<div class="field"><label class="label">Genre</label><div class="select"><select @change=${this.genreChanged} ?disabled=${readOnly} id="movieGenreId">${templates}</select></div></div>`;
    }

    hidden(id, value) {
        return html`<input type="hidden" value=${value} id=${id} />`;
    }

    focus() {
        const focused = document.querySelector(".is-focused");
        if (focused) {
            focused.focus();
        }
    }

    title(title) {
        return html`<h1 class="title">${title}</h1>`;
    }

    id(readOnly, value) {
        var disabled = readOnly ? "New" : "";
        return html`
        <div class="field">
        <label class="label">ID</label>
        <div class="control has-icons-left">
          <input id="movieId" class="input" type="text" placeholder="${disabled}" disabled value="${value}">
          <span class="icon is-small is-left">
            <i class="fas fa-hashtag"></i>
          </span>
        </div>
      </div>`;
    }

    text(id, label, readOnly, value) {
        return html`<div class="field">
        <label class="label">${label}</label>
        <div class="control has-icons-left">
          <input id="${id}" class="input" ?disabled="${readOnly}" type="text" minlength="2" value="${value}">
          <span class="icon is-small is-left">
            <i class="fas fa-font"></i>
          </span>
        </div>
      </div>`;
    }

    textArea(id, label, readOnly, value) {
        return html`<div class="field">
        <label class="label">${label}</label>
        <div class="control">
          <textarea id=${id} class="textarea" ?disabled="${readOnly}">${value}</textarea>
        </div>
      </div>`;
    }

    number(id, label, readOnly, value) {
        return html`<div class="field">
        <label class="label">${label}</label>
        <div class="control has-icons-left">
          <input id="${id}" class="input" ?disabled="${readOnly}" type="number" min="0" placeholder="Display order" value="${value}">
          <span class="icon is-small is-left">
            <i class="fas fa-sort-numeric-down"></i>
          </span>
        </div>
      </div>`;
    }

    showSuccess() {
        document.querySelector("#successMessage").classList.remove("is-hidden");
    }

    showError(err) {
        document.querySelector("#errorMessage").classList.remove("is-hidden");
        document.querySelector("#errorMessage>div>p").innerText = err.ok === false
            ? (err.status + " " + err.statusText)
            : err;
    }

    genreChanged(event) {
        var currentGenre = document.querySelector("#movieGenreId").value;
        var series = document.querySelector("#movieSerieId");
        series.options.length = 0;
        for (const [serieId, genreId] of Object.entries(this.data.serieGenres)) {
            if (genreId == currentGenre) {
                series.options[series.options.length] = new Option(this.data.series[serieId], serieId);
            }
        }
    }

    initUploadListener() {
        const fileInput = document.querySelector('#thumbnail input[type=file]');
        const id = document.querySelector("#movieId").value;
        fileInput.onchange = () => {
            if (fileInput.files.length > 0) {
                document.querySelector("#thumbnailProgress").classList.remove("is-hidden");

                const file = fileInput.files[0];
                const fileName = file.name;
                const fileNameInput = document.querySelector('#thumbnail .file-name');
                fileNameInput.textContent = fileName;
                new FileUpload(file, file.file, "genres", id);
            }
        }
    }

    selectItunesThumbnailClicked(event) {
        this.itunesCallback(document.querySelector("#movieName").value);
    }

    selectItunesThumbnailSelectClicked(event) {
        var anchor = event.target.closest("a");
        var id = anchor.getAttribute("data-id");
        var url = this.itunesData[id].artworkUrl;
        var dataUrl = encodeURIComponent(btoa(url));
        render(html`<center><figure data-url="${dataUrl}" class="image is-128x128"><img src="${url}"></figure></center>`, document.querySelector("#itunesDetail"));
        document.querySelector("#itunesDetailTitle").innerText = this.itunesData[id].trackName;
        document.querySelector("#itunesModalDetail").classList.toggle("is-active");
    }

    selectItunesThumbnailCloseClicked() {
        document.querySelector("#itunesModal").classList.toggle("is-active");
    }

    selectItunesEmptyThumbnailCloseClicked() {
        document.querySelector("#itunesModalEmpty").classList.toggle("is-active");
    }

    selectItunesDetailThumbnailCloseClicked() {
        document.querySelector("#itunesModalDetail").classList.toggle("is-active");
    }

    selectItunesDetailThumbnailSaveClicked() {
        var figure = document.querySelector("#itunesDetail>figure");
        console.log("selectItunesDetailThumbnailSaveClicked", figure);
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
            document.querySelector("#itunes_" + i).addEventListener("click", event => {this.selectItunesThumbnailSelectClicked(event)})
        }
        document.querySelector("#itunesModal").classList.toggle("is-active");
    }

    setItunesCallback(itunesCallback) {
        this.itunesCallback = itunesCallback;
    }

    selectItunesThumbnailModal(readOnly) {
        if (readOnly) {
            return html``;
        }

        var button = html`<button type="button" class="button is-link is-light" @click="${this.selectItunesThumbnailClicked}">iTunes Thumbnail</button>`;
        var listOfArtwork = html`<div class="modal" id="itunesModal">
            <div class="modal-background"></div>
            <div class="modal-card">
            <header class="modal-card-head">
                <p class="modal-card-title">iTunes Artwork</p>
                <button class="delete" aria-label="close" type="button" @click="${this.selectItunesThumbnailCloseClicked}"></button>
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
                <button class="delete" aria-label="close" type="button" @click="${this.selectItunesEmptyThumbnailCloseClicked}"></button>
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
                <button class="delete" aria-label="close" type="button" @click="${this.selectItunesDetailThumbnailCloseClicked}"></button>
            </header>
            <section class="modal-card-body">
                <div id="itunesDetail" style="height:230px;"></div>
            </section>
            <footer class="modal-card-foot">
                <button type="button" class="button is-success" @click="${this.selectItunesDetailThumbnailSaveClicked}">Save</button>
                <button type="button" class="button" @click="${this.selectItunesDetailThumbnailCloseClicked}">Cancel</button>
            </footer>
            </div>
        </div>`;

        return html`${button} ${listOfArtwork} ${emptyResult} ${artworkDetail}`;
    }
}
