import { html, render } from "lit-html";
import { MovieService } from "/_dist_/movie/MovieService.js";
import { BreadCrumbs } from "/_dist_/BreadCrumbs.js";

export class MovieView extends HTMLElement {
    constructor() {
        super();
        this.data = [];
        this.service = new MovieService();
    }

    connectedCallback() {
        addEventListener(this.service.getCustomEventName(), e => this.update(e));
        console.log("MovieView.connectedCallback");
        this.render();
    }

    update(event) {
        console.log("MovieView.update: Received event ", event);

        const e = new CustomEvent(BreadCrumbs.eventName(), [{url: "/", title: "Home"}]);
        this.dispatchEvent(e);
        console.log("MovieView.update: Dispatched event ", e);

        this.data = this.service.getData(event.detail.key);

        render(this.view(), this);
    }

    render() {
        var id = this.location.params.id;
        console.log("MovieView.render: running fetchData(" + id + ")");
        this.service.setId(id);
        this.service.fetchData();
    }

    view() {
        const name = this.renderName();
        const moviePlayer = this.renderMoviePlayer();
        const details = this.renderDetails();
        return html`${name} ${moviePlayer} ${details}`;
    }

    renderName() {
        return html`
            <div class="container">
            <div class="columns is-centered"><div class="column is-half">
            <h1 class="title">
              ${this.data.name}
            </h1>
            <h2 class="subtitle">
              ${this.data.year}
            </h2>
            </div></div>
`;
    }

    renderMoviePlayer() {
        return html`<div class="columns is-centered"><div class="column is-half"><video controls src="${this.data.dto.url}" poster="${this.data.dto.thumbnail}">
        <a href="${this.data.dto.url}">
            <img src="${this.data.dto.thumbnail}" alt="${this.data.dto.title}"></img>
        </a>
    </video>
    </div></div></div>`;
    }

    renderDetails() {
        var desc;
        if (null === this.data.description || undefined === this.data.description || this.data.description.trim().length === 0) {
            desc = "";
        } else {
            desc = html`<div class="card"><div class="card-content"><div class="content">${this.data.description}</div></div></div>`;
        }

        return html`<div class="columns is-centered">
        <div class="column">&nbsp;</div>
        <div class="column is-half">
        ${desc}
        </div>
        <div class="column">&nbsp;</div>
        </div>

        <div class="columns is-centered">
        <div class="column">&nbsp;</div>
        <div class="column">
            <div class="field">
                <label class="label">Genre</label>
                <div class="control">
                    <input class="input" type="text" readonly value="${this.data.genreName}">
                </div>
            </div>
        </div>
        <div class="column">
            <div class="field">
                <label class="label">Serie</label>
                <div class="control">
                    <input class="input" type="text" readonly value="${this.data.serieName}">
                </div>
            </div>
        </div>
        <div class="column">&nbsp;</div>
        </div>

        <div class="columns is-centered">
        <div class="column">&nbsp;</div>
        <div class="column">
            <div class="field">
                <label class="label">Language</label>
                <div class="control">
                    <input class="input" type="text" readonly value="${this.data.languageName}">
                </div>
            </div>
        </div>
        <div class="column">
            <div class="field">
                <label class="label">Original Language</label>
                <div class="control">
                    <input class="input" type="text" readonly value="${this.data.originalLanguageName}">
                </div>
            </div>
        </div>
        <div class="column">&nbsp;</div>
        </div>

        <div class="columns is-centered">
        <div class="column">&nbsp;</div>
        <div class="column">
            <div class="field">
                <label class="label">Subtitles</label>
                <div class="control">
                    <input class="input" type="text" readonly value="${this.data.subtitlesName}">
                </div>
            </div>
        </div>
        <div class="column">&nbsp;</div>
        <div class="column">&nbsp;</div>
        </div>`;
    }
}

customElements.define("movie-view", MovieView);
