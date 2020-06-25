import { html } from "lit-html";

export class RenderMovieDetail {
    constructor(successMessage, errorMessage) {
        this.successMessage = successMessage;
        this.errorMessage = errorMessage;
        this.serieChanged = this.serieChanged.bind(this);
    }

    view(entity, readOnly, isNew) {
        let items = [];

        items.push(this.message(this.successMessage, "is-primary", "successMessage"));
        items.push(this.message(this.errorMessage, "is-danger", "errorMessage"));

        if (! isNew) {
            items.push(this.id(readOnly, entity.id));
        }

        console.log("Got Movie Entity", entity);

        items.push(this.text("movieName", "Name", readOnly, entity.name));
        items.push(this.text("movieOriginalName", "Original Name", readOnly, entity.originalName));
        items.push(this.textArea("movieDescription", "Description", readOnly, entity.description));
        items.push(this.combo("movieGenreId", "Genre", readOnly, entity.genreId, entity.genres, this.dummy));
        items.push(this.combo("movieSerieId", "Serie", readOnly, entity.serieId, entity.series, this.serieChanged));
        items.push(this.combo("movieLanguageId", "Language", readOnly, entity.languageId, entity.languages));
        items.push(this.combo("movieOriginalLanguageId", "OriginalLanguage", readOnly, entity.originalLanguageId, entity.languages));
        items.push(this.combo("movieSubtitlesId", "Subtitles", readOnly, entity.subtitlesId, entity.languages));
        items.push(this.number("movieDisplayOrder", "Display Order", readOnly, entity.displayOrder));

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

    combo(id, label, readOnly, itemId, items, callback) {
        var templates = [];
        for (const [key, value] of Object.entries(items)) {
            templates.push(html`<option value="${key}" ?selected=${key == itemId}>${value}</option>`);
        }

        return html`<div class="field"><label class="label">${label}</label><div class="select"><select @onchange=${callback} id="${id}" ?disabled=${readOnly}>${templates}</select></div></div>`;
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

    serieChanged(event) {
        console.log("RenderMovieDetail.serieChanged", event);
        // ... serieGenres: serieId -> genreId
    }
}
