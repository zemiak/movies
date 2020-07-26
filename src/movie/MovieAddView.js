import { html, render } from "lit-html";
import { RenderMovieDetail } from "/_dist_/movie/RenderMovieDetail.js";
import { MovieDetailService } from "/_dist_/movie/MovieDetailService.js";
import { MovieAddService } from "/_dist_/movie/MovieAddService.js";

export class MovieAddView extends HTMLElement {
    constructor() {
        super();
        this.renderer = new RenderMovieDetail("Successfully added", "Add error");
        this.service = new MovieDetailService();
        this.addservice = new MovieAddService();
        this.saveClick = this.saveClick.bind(this);
    }

    connectedCallback() {
        addEventListener(this.addservice.getCustomEventName(), e => this.update(e));
        this.render();
    }

    update() {
        this.data = this.addservice.getData(event.detail.key);
        render(this.view(), this);
        this.renderer.initUploadListener();
        this.renderer.focus();
    }

    render() {
        this.addservice.fetchData();
    }

    view() {
        var view = this.renderer.view(this.data, false, true);
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
        console.log("MovieAddView.saveClick - sending data", item);
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

        console.log("MovieAddView.saveSuccess", response);
        this.renderer.showSuccess();
        setTimeout(_ => window.location = "/admin/movies", 1500);
    }

    saveError(err) {
        console.log("MovieAddView.saveError", err);
        this.renderer.showError(err);
        setTimeout(_ => window.location = "/admin/movies", 3000);
    }

    cancelClick(event) {
        window.location = "/admin/movies";
    }
}

customElements.define("movie-add-view", MovieAddView);
