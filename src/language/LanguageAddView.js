import { html, render } from "lit-html";
import { RenderLanguageDetail } from "/_dist_/language/RenderLanguageDetail.js";

export class LanguageAddView extends HTMLElement {
    constructor() {
        super();
        this.renderer = new RenderLanguageDetail("Successfully added", "Add error");
    }

    connectedCallback() {
        this.render();
        this.renderer.focus();
    }

    render() {
        render(this.view(), this);
    }

    view() {
        var entity = {id: "", code: "", name: "", displayOrder: 0};
        var view = this.renderer.view(entity, false, true);
        let buttons = this.buttons();
        return html`${view}${buttons}`;
    }

    buttons(readOnly, isNew) {
        return html`<div class="field is-grouped">
        <div class="control">
          <button class="button is-link" @click="${this.addClick}">Add</button>
        </div>
        <div class="control">
          <button class="button is-link is-light" @click="${this.cancelClick}">Cancel</button>
        </div>
      </div>
      `;
    }

    addClick(event) {
        console.log("LanguageEditView.saveClick", event);
    }

    cancelClick(event) {
        window.location = "/admin/languages";
    }

    addSuccess(event) {

    }
}

customElements.define("language-add-view", LanguageAddView);
