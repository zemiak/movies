import { html, render } from "lit-html";
import { RenderSerieDetail } from "/_dist_/serie/RenderSerieDetail.js";
import { SerieDetailService } from "/_dist_/serie/SerieDetailService.js";
import { SerieAddService } from "/_dist_/serie/SerieAddService.js";

export class SerieAddView extends HTMLElement {
    constructor() {
        super();
        this.renderer = new RenderSerieDetail("Successfully added", "Add error");
        this.service = new SerieDetailService();
        this.addservice = new SerieAddService();
        this.saveClick = this.saveClick.bind(this);
    }

    connectedCallback() {
        addEventListener(this.addservice.getCustomEventName(), e => this.update(e));
        console.log("SerieAddView.connectedCallback");
        this.render();
    }

    update() {
        this.data = this.addservice.getData(event.detail.key);
        render(this.view(), this);
        this.renderer.focus();
    }

    render() {
        this.addservice.fetchData();
    }

    view() {
        var entity = {id: "", code: "", name: "", displayOrder: 0, genres: this.data.genres, tvShow: "0"};
        var view = this.renderer.view(entity, false, true);
        let buttons = this.buttons();
        return html`${view}${buttons}`;
    }

    buttons(readOnly, isNew) {
        return html`<div class="field is-grouped">
        <div class="control">
          <button id="saveButton" class="button is-link" @click="${this.saveClick}">Add</button>
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
        item.id = null;
        console.log("SerieAddView.saveClick - sending data", item);
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

        console.log("SerieAddView.saveSuccess", response);
        this.renderer.showSuccess();
        setTimeout(_ => window.location = "/admin/series", 1500);
    }

    saveError(err) {
        console.log("SerieAddView.saveError", err);
        this.renderer.showError(err);
        setTimeout(_ => window.location = "/admin/series", 3000);
    }

    cancelClick(event) {
        window.location = "/admin/series";
    }
}

customElements.define("serie-add-view", SerieAddView);
