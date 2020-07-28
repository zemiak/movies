import { html } from "lit-html";
import { FileUpload } from "/_dist_/files/FileUpload.js";

export class RenderSerieDetail {
    constructor(successMessage, errorMessage) {
        this.successMessage = successMessage;
        this.errorMessage = errorMessage;
    }

    view(entity, readOnly, isNew) {
        let items = [];

        items.push(this.message(this.successMessage, "is-primary", "successMessage"));
        items.push(this.message(this.errorMessage, "is-danger", "errorMessage"));

        if (! isNew) {
            items.push(this.id(readOnly, entity.id));
        }

        items.push(this.name(readOnly, entity.name));
        items.push(this.genre(readOnly, entity.genreId, entity.genres));
        items.push(this.displayOrder(readOnly, entity.displayOrder));
        items.push(this.tvShow(readOnly, entity.tvShow));
        items.push(this.hidden("seriePictureFileName", entity.pictureFileName));
        items.push(this.hidden("serieCreated", entity.created));
        items.push(this.thumbnail(readOnly, entity.thumbnailUrl));

        let title = this.title(isNew ? "New Serie" : "Serie");

        return html`${title}<form class="form-horizontal"><fieldset>${items}</fieldset></form><p>&nbsp;</p>`;
    }

    getFormData() {
        var item = {};

        this.addItem(item, "id", "serieId");
        this.addItem(item, "name", "serieName");
        this.addItem(item, "genreId", "serieGenreId");
        this.addItem(item, "displayOrder", "serieDisplayOrder");
        item["tvShow"] = document.querySelector("#serieTvShow").checked ? true : false;
        this.addItem(item, "pictureFileName", "seriePictureFileName");
        this.addItem(item, "created", "serieCreated");

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

    genre(readOnly, genreId, genres) {
        var items = [];
        for (const [key, value] of Object.entries(genres)) {
            items.push(html`<option value="${key}" ?selected=${key == genreId}>${value}</option>`);
        }

        return html`<div class="field"><label class="label">Genre</label><div class="select"><select id="serieGenreId" ?disabled=${readOnly}>${items}</select></div></div>`;
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
          <input id="serieId" class="input" type="text" placeholder="${disabled}" disabled value="${value}">
          <span class="icon is-small is-left">
            <i class="fas fa-hashtag"></i>
          </span>
        </div>
      </div>`;
    }

    hidden(id, value) {
        return html`<input type="hidden" value=${value} id=${id} />`;
    }

    tvShow(readOnly, value) {
        return html`<div class="field">
        <label class="checkbox">
        <input id="serieTvShow" ?disabled="${readOnly}" type="checkbox" ?checked="${value}">
            TV Show
        </label>
      </div>`;
    }

    name(readOnly, value) {
        return html`<div class="field">
        <label class="label">Description</label>
        <div class="control has-icons-left">
          <input id="serieName" class="input" ?disabled="${readOnly}" type="text" minlength="2" placeholder="Serie description" value="${value}">
          <span class="icon is-small is-left">
            <i class="fas fa-font"></i>
          </span>
        </div>
      </div>`;
    }

    displayOrder(readOnly, value) {
        return html`<div class="field">
        <label class="label">Order</label>
        <div class="control has-icons-left">
          <input id="serieDisplayOrder" class="input" ?disabled="${readOnly}" type="number" min="0" placeholder="Display order" value="${value}">
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

    initUploadListener() {
        const fileInput = document.querySelector('#thumbnail input[type=file]');
        const id = document.querySelector("#serieId").value;
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
}
