import { html, render } from "lit-html";
import { RenderLanguageDetail } from "/_dist_/language/RenderLanguageDetail.js";
import { LanguageDetailService } from "/_dist_/language/LanguageDetailService.js";

export class LanguageEditView extends HTMLElement {
    constructor() {
        super();
        this.renderer = new RenderLanguageDetail();
        this.service = new LanguageDetailService();
    }

    connectedCallback() {
        addEventListener(this.service.getCustomEventName(), e => this.update(e));
        console.log("LanguageEditView.connectedCallback");
        this.render();
    }

    update() {
        this.data = this.service.getData(event.detail.key);
        render(this.view(), this);
        this.renderer.focus();
    }

    render() {
        var id = this.location.params.id;
        this.service.setId(id);
        this.service.fetchData();
    }

    view() {
        var view = this.renderer.view(this.data, false, false);
        let buttons = this.buttons();
        return html`${view}${buttons}`;
    }

    buttons() {
        return html`<div class="field is-grouped">
        <div class="control">
          <button class="button is-link" @click="${this.saveClick}">Save</button>
        </div>
        <div class="control">
          <button class="button is-link is-light" @click="${this.cancelClick}">Cancel</button>
        </div>
      </div>
      `;
    }

    saveClick(event) {
        console.log("LanguageEditView.saveClick", event);
    }

    cancelClick(event) {
        console.log("LanguageEditView.cancelClick", event);
    }
}

customElements.define("language-edit-view", LanguageEditView);
