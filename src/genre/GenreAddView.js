import { html, render } from "lit-html";
import { RenderGenreDetail } from "/_dist_/genre/RenderGenreDetail.js";
import { GenreDetailService } from "/_dist_/genre/GenreDetailService.js";

export class GenreAddView extends HTMLElement {
    constructor() {
        super();
        this.renderer = new RenderGenreDetail("Successfully added", "Add error");
        this.service = new GenreDetailService();
        this.saveClick = this.saveClick.bind(this);
    }

    connectedCallback() {
        this.render();
        this.renderer.focus();
    }

    render() {
        render(this.view(), this);
    }

    view() {
        var entity = {id: "", code: "", name: "", displayOrder: 0, protectedGenre: "0"};
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
        console.log("GenreAddView.saveClick - sending data", item);
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

        console.log("GenreAddView.saveSuccess", response);
        this.renderer.showSuccess();
        setTimeout(_ => window.location = "/admin/genres", 1500);
    }

    saveError(err) {
        console.log("GenreAddView.saveError", err);
        this.renderer.showError(err);
        setTimeout(_ => window.location = "/admin/genres", 3000);
    }

    cancelClick(event) {
        window.location = "/admin/genres";
    }
}

customElements.define("genre-add-view", GenreAddView);
