import { html } from "lit-html";

export class RenderGenreDetail {
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
        items.push(this.displayOrder(readOnly, entity.displayOrder));
        items.push(this.protectedGenre(readOnly, entity.protectedGenre));
        items.push(this.hidden("genrePictureFileName", entity.pictureFileName));
        items.push(this.hidden("genreCreated", entity.created));
        items.push(this.thumbnail(readOnly, entity.thumbnailUrl));

        let title = this.title(isNew ? "New Genre" : "Genre");

        return html`${title}<form class="form-horizontal"><fieldset>${items}</fieldset></form><p>&nbsp;</p>`;
    }

    getFormData() {
        var item = {};

        this.addItem(item, "id", "genreId");
        this.addItem(item, "name", "genreName");
        this.addItem(item, "displayOrder", "genreDisplayOrder");
        item["protectedGenre"] = document.querySelector("#genreProtectedGenre").checked ? "1" : "0";
        this.addItem(item, "pictureFileName", "genrePictureFileName");
        this.addItem(item, "created", "genreCreated");

        // {id: "", name: "", displayOrder: 0, protectedGenre: "0", created: null, pictureFileName: null};


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
          <input class="file-input" type="file" name="resume" id="thumbnail">
          <span class="file-cta">
            <span class="file-icon">
              <i class="fas fa-upload"></i>
            </span>
            <span class="file-label">
              Choose a file…
            </span>
          </span>
          <span class="file-name">
      No file uploaded
    </span>
        </label>
      </div>`;
        return html`<figure class="image is-128x128">
        <img class="is-rounded" src="${imageUrl}">
      </figure>${upload}`;
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
          <input id="genreId" class="input" type="text" placeholder="${disabled}" disabled value="${value}">
          <span class="icon is-small is-left">
            <i class="fas fa-hashtag"></i>
          </span>
        </div>
      </div>`;
    }

    protectedGenre(readOnly, intValue) {
        var value = intValue === 1;
        return html`<div class="field">
        <label class="checkbox">
        <input id="genreProtectedGenre" ?disabled="${readOnly}" type="checkbox" ?checked="${value}">
            Protected
        </label>
      </div>`;
    }

    name(readOnly, value) {
        return html`<div class="field">
        <label class="label">Description</label>
        <div class="control has-icons-left">
          <input id="genreName" class="input" ?disabled="${readOnly}" type="text" minlength="2" placeholder="Genre description" value="${value}">
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
          <input id="genreDisplayOrder" class="input" ?disabled="${readOnly}" type="number" min="0" placeholder="Display order" value="${value}">
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
        fileInput.onchange = () => {
            if (fileInput.files.length > 0) {
                const fileName = document.querySelector('#thumbnail .file-name');
                fileName.textContent = fileInput.files[0].name;
            }
        }
    }
}
