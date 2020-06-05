export class RenderLanguageDetail {
    render(entity, readOnly, isNew) {
        return this.id(readOnly) + this.code(readOnly) + this.name(readOnly) + this.displayOrder(readOnly)
            + this.buttons(readOnly, isNew);
    }

    id(readOnly, value) {
        var disabled = readOnly ? "New" : "";
        return `<fieldset disabled>
        <div class="field">
        <label class="label">ID</label>
        <div class="control has-icons-left">
          <input class="input is-success" type="text" placeholder="${disabled}" value="${value}">
          <span class="icon is-small is-left">
            <i class="fas fa-hashtag"></i>
          </span>
        </div>
      </div></fieldset>`;
    }

    code(readOnly, value) {
        var disabled = readOnly ? "disabled" : "";
        return `<fieldset ${disabled}>
        <div class="field">
        <label class="label">Code</label>
        <div class="control has-icons-left">
          <input class="input is-success" type="text" placeholder="Language code" minlength="2" maxlength="2" value="${value}">
          <span class="icon is-small is-left">
            <i class="fas fa-globe"></i>
          </span>
        </div>
      </div></fieldset>`;
    }

    name(readOnly, value) {
        var disabled = readOnly ? "disabled" : "";
        return `<fieldset ${disabled}>
        <div class="field">
        <label class="label">Description</label>
        <div class="control has-icons-left">
          <input class="input is-success" type="text" minlength="2" placeholder="Language description" value="${value}">
          <span class="icon is-small is-left">
            <i class="fas fa-font"></i>
          </span>
        </div>
      </div></fieldset>`;
    }

    displayOrder(readOnly, value) {
        var disabled = readOnly ? "disabled" : "";
        return `<fieldset ${disabled}>
        <div class="field">
        <label class="label">Order</label>
        <div class="control has-icons-left">
          <input class="input is-success" type="number" min="0" placeholder="Display order" value="${value}">
          <span class="icon is-small is-left">
            <i class="fas fa-sort-numeric-down"></i>
          </span>
        </div>
      </div></fieldset>`;
    }

    buttons(readOnly, isNew) {
        return `<div class="field is-grouped">
        <div class="control">
          <button class="button is-link">Submit</button>
        </div>
        <div class="control">
          <button class="button is-link is-light">Cancel</button>
        </div>
      </div>
      `;
    }
}
