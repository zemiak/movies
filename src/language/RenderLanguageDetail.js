import { html } from "lit-html";

export class RenderLanguageDetail {
    view(entity, readOnly, isNew) {
        let items = [];
        if (! isNew) {
            items.push(this.id(readOnly, entity.id));
        }

        items.push(this.code(readOnly, entity.code));
        items.push(this.name(readOnly, entity.name));
        items.push(this.displayOrder(readOnly, entity.displayOrder));

        let title = this.title(isNew ? "New Language" : "Language");

        return html`${title}<p>&nbsp;</p><form class="form-horizontal"><fieldset>${items}</fieldset></form><p>&nbsp;</p>`;
    }

    focus() {
        const focused = document.querySelector(".is-focused");
        if (focused) {
            focused.focus();
        }
    }

    title(title) {
        return html`<section class="hero">
        <div class="">
          <div class="container">
            <h1 class="title">
              ${title}
            </h1>
          </div>
        </div>
      </section>`;
    }

    id(readOnly, value) {
        var disabled = readOnly ? "New" : "";
        return html`
        <div class="field">
        <label class="label">ID</label>
        <div class="control has-icons-left">
          <input class="input" type="text" placeholder="${disabled}" disabled value="${value}">
          <span class="icon is-small is-left">
            <i class="fas fa-hashtag"></i>
          </span>
        </div>
      </div>`;
    }

    code(readOnly, value) {
        var disabled = readOnly ? "disabled" : "";
        return html`<div class="field">
        <label class="label">Code</label>
        <div class="control has-icons-left">
          <input class="input is-focused" ?disabled="${readOnly}" type="text" placeholder="Language code" minlength="2" maxlength="2" value="${value}">
          <span class="icon is-small is-left">
            <i class="fas fa-globe"></i>
          </span>
        </div>
      </div>`;
    }

    name(readOnly, value) {
        var disabled = readOnly ? "disabled" : "";
        return html`<div class="field">
        <label class="label">Description</label>
        <div class="control has-icons-left">
          <input class="input" ?disabled="${readOnly}" type="text" minlength="2" placeholder="Language description" value="${value}">
          <span class="icon is-small is-left">
            <i class="fas fa-font"></i>
          </span>
        </div>
      </div>`;
    }

    displayOrder(readOnly, value) {
        var disabled = readOnly ? "disabled" : "";
        return html`<div class="field">
        <label class="label">Order</label>
        <div class="control has-icons-left">
          <input class="input" ?disabled="${readOnly}" type="number" min="0" placeholder="Display order" value="${value}">
          <span class="icon is-small is-left">
            <i class="fas fa-sort-numeric-down"></i>
          </span>
        </div>
      </div>`;
    }
}
