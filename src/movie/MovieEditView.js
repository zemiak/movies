import { html, render } from "lit-html";
import { RenderMovieDetail } from "/_dist_/movie/RenderMovieDetail.js";
import { MovieDetailService } from "/_dist_/movie/MovieDetailService.js";

export class MovieEditView extends HTMLElement {
    constructor() {
        super();
        this.renderer = new RenderMovieDetail("Successfully saved", "Save error");
        this.service = new MovieDetailService();
        this.saveClick = this.saveClick.bind(this);
    }

    connectedCallback() {
        addEventListener(this.service.getCustomEventName(), e => this.update(e));
        this.render();
    }

    update() {
        this.data = this.service.getData(event.detail.key);
        render(this.view(), this);
        this.renderer.initUploadListener();
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
        console.log("MovieEditView.saveClick - sending data", item);
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

        console.log("MovieEditView.saveSuccess", response);
        this.renderer.showSuccess();
        setTimeout(_ => window.location = "/admin/movies", 1500);
    }

    saveError(err) {
        console.log("MovieEditView.saveError", err);
        this.renderer.showError(err);
        setTimeout(_ => window.location = "/admin/movies", 3000);
    }

    cancelClick(event) {
        window.location = "/admin/movies";
    }
}

customElements.define("movie-edit-view", MovieEditView);
