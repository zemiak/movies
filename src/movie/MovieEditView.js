import { html, render } from "lit-html";
import { RenderMovieDetail } from "/_dist_/movie/RenderMovieDetail.js";
import { MovieDetailService } from "/_dist_/movie/MovieDetailService.js";
import { MovieItunesService } from "/_dist_/movie/itunes/MovieItunesService.js";
import { MovieCsfdService } from "/_dist_/movie/csfd/MovieCsfdService.js";
import { MovieImdbService } from "/_dist_/movie/imdb/MovieImdbService.js";

export class MovieEditView extends HTMLElement {
    constructor() {
        super();
        this.renderer = new RenderMovieDetail("Successfully saved", "Save error",
            {
                "itunes": {
                    "fetch": (data) => {this.fetchItunesThumbnails(data)},
                    "save": (data) => {this.saveItunesThumbnails(data)}
                },

                "csfd": {
                    "fetch": (data) => {this.fetchCsfdThumbnails(data)},
                    "save": (data) => {this.saveCsfdThumbnails(data)}
                },

                "imdb": {
                    "fetch": (data) => {this.fetchImdbThumbnails(data)},
                    "save": (data) => {this.saveImdbThumbnails(data)}
                }
            }
        );
        this.service = new MovieDetailService();
        this.itunesService = new MovieItunesService();
        this.csfdService = new MovieCsfdService();
        this.imdbService = new MovieImdbService();
        this.saveClick = this.saveClick.bind(this);
    }

    connectedCallback() {
        addEventListener(this.service.getCustomEventName(), e => this.update(e));
        addEventListener(this.itunesService.getCustomEventName(), e => this.updateItunesThumbnails(e));
        addEventListener(this.csfdService.getCustomEventName(), e => this.updateCsfdThumbnails(e));
        addEventListener(this.imdbService.getCustomEventName(), e => this.updateImdbThumbnails(e));
        this.render();
    }

    update() {
        this.data = this.service.getData(event.detail.key);
        render(this.view(), this);
        this.renderer.initUploadListener();
        this.renderer.focus();
    }

    updateItunesThumbnails() {
        this.renderer.updateItunesThumbnails(this.itunesService.getData(event.detail.key));
    }

    fetchItunesThumbnails(name) {
        this.itunesService.setName(name);
        this.itunesService.fetchData();
    }

    saveItunesThumbnails(url) {
        var id = this.location.params.id;
        this.itunesService.saveThumbnail(id, url,
            response => this.saveSuccessItunesThumbnail(response),
            err => this.saveErrorItunesThumbnail(err));
    }

    updateCsfdThumbnails() {
        this.renderer.updateCsfdThumbnails(this.csfdService.getData(event.detail.key));
    }

    fetchCsfdThumbnails(name) {
        this.csfdService.setName(name);
        this.csfdService.fetchData();
    }

    saveCsfdThumbnails(data) {
        var id = this.location.params.id;
        this.csfdService.saveThumbnail(id, data,
            response => this.saveSuccessCsfdThumbnail(response),
            err => this.saveErrorCsfdThumbnail(err));
    }

    updateImdbThumbnails() {
        this.renderer.updateImdbThumbnails(this.imdbService.getData(event.detail.key));
    }

    fetchImdbThumbnails(name) {
        this.imdbService.setName(name);
        this.imdbService.fetchData();
    }

    saveImdbThumbnails(data) {
        var id = this.location.params.id;
        this.imdbService.saveThumbnail(id, data,
            response => this.saveSuccessImdbThumbnail(response),
            err => this.saveErrorImdbThumbnail(err));
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

        this.renderer.showSuccess();
        setTimeout(_ => window.location = "/admin/movies", 1500);
    }

    saveError(err) {
        this.renderer.showError(err);
        setTimeout(_ => window.location = "/admin/movies", 3000);
    }

    saveSuccessItunesThumbnail(response) {
        if (! response.ok) {
            this.saveErrorItunesThumbnail(response);
            return;
        }

        response.json().then(data => {
            document.querySelector("#moviePictureFileName").value = data.pictureFileName;
        });

        this.renderer.showSuccessItunesThumbnail(response);
    }

    saveErrorItunesThumbnail(err) {
        this.renderer.showErrorItunesThumbnail(err);
    }

    saveSuccessCsfdThumbnail(response) {
        if (! response.ok) {
            this.saveErrorCsfdThumbnail(response);
            return;
        }

        response.json().then(data => {
            document.querySelector("#moviePictureFileName").value = data.pictureFileName;
            document.querySelector("#movieDescription").value = data.description;
            document.querySelector("#movieYear").value = data.year;
        });

        this.renderer.showSuccessCsfdThumbnail(response);
    }

    saveErrorCsfdThumbnail(err) {
        this.renderer.showErrorCsfdThumbnail(err);
    }

    saveSuccessImdbThumbnail(response) {
        if (! response.ok) {
            this.saveErrorImdbThumbnail(response);
            return;
        }

        response.json().then(data => {
            document.querySelector("#moviePictureFileName").value = data.pictureFileName;
            document.querySelector("#movieDescription").value = data.description;
            document.querySelector("#movieYear").value = data.year;
        });

        this.renderer.showSuccessImdbThumbnail(response);
    }

    saveErrorImdbThumbnail(err) {
        this.renderer.showErrorImdbThumbnail(err);
    }

    cancelClick(event) {
        window.location = "/admin/movies";
    }
}

customElements.define("movie-edit-view", MovieEditView);
