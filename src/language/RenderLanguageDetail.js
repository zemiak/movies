import { html } from "lit-html";

export class RenderLanguageDetail {
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

        items.push(this.code(readOnly, entity.code));
        items.push(this.name(readOnly, entity.name));
        items.push(this.displayOrder(readOnly, entity.displayOrder));

        let title = this.title(isNew ? "New Language" : "Language");

        return html`${title}<form class="form-horizontal"><fieldset>${items}</fieldset></form><p>&nbsp;</p>`;
    }

    getFormData() {
        var item = {};

        this.addItem(item, "id", "languageId");
        this.addItem(item, "code", "languageCode");
        this.addItem(item, "name", "languageName");
        this.addItem(item, "displayOrder", "languageDisplayOrder");

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
          <input id="languageId" class="input" type="text" placeholder="${disabled}" disabled value="${value}">
          <span class="icon is-small is-left">
            <i class="fas fa-hashtag"></i>
          </span>
        </div>
      </div>`;
    }

    code(readOnly, value) {
        return html`<div class="field">
        <label class="label">Code</label>
        <div class="control has-icons-left">
          <input id="languageCode" class="input is-focused" ?disabled="${readOnly}" type="text" placeholder="Language code" minlength="2" maxlength="2" value="${value}">
          <span class="icon is-small is-left">
            <i class="fas fa-globe"></i>
          </span>
        </div>
      </div>`;
    }

    name(readOnly, value) {
        return html`<div class="field">
        <label class="label">Description</label>
        <div class="control has-icons-left">
          <input id="languageName" class="input" ?disabled="${readOnly}" type="text" minlength="2" placeholder="Language description" value="${value}">
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
          <input id="languageDisplayOrder" class="input" ?disabled="${readOnly}" type="number" min="0" placeholder="Display order" value="${value}">
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
}
