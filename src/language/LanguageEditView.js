import { html, render } from "lit-html";
import { RenderLanguageDetail } from "/_dist_/language/RenderLanguageDetail.js";
import { LanguageDetailService } from "/_dist_/language/LanguageDetailService.js";

export class LanguageEditView extends HTMLElement {
    constructor() {
        super();
        this.renderer = new RenderLanguageDetail("Successfully saved", "Save error");
        this.service = new LanguageDetailService();

        this.saveClick = this.saveClick.bind(this);
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
          <button id="saveButton" class="button is-link" @click="${this.saveClick}">Save</button>
        </div>
        <div class="control">
          <button class="button is-link is-light" @click="${this.cancelClick}">Cancel</button>
        </div>
      </div>
      `;
    }

    saveClick(event) {
        document.querySelector("#saveButton").disabled = "disabled";
        var item = this.renderer.getFormData();
        console.log("LanguageEditView.saveClick - sending data", item);
        this.service.saveOrUpdate(item,
            response => this.saveSuccess(response),
            err => this.saveError(err)
        );
    }

    saveSuccess(response) {
        if (! response.ok) {
            this.saveError(response);
            return;
        }

        console.log("LanguageEditView.saveSuccess", response);
        this.renderer.showSuccess();
        setTimeout(_ => window.location = "/admin/languages", 1500);
    }

    saveError(err) {
        console.log("LanguageEditView.saveError", err);
        this.renderer.showError(err);
        setTimeout(_ => window.location = "/admin/languages", 3000);
    }

    cancelClick(event) {
        window.location = "/admin/languages";
    }
}

customElements.define("language-edit-view", LanguageEditView);
