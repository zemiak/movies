import { html, render } from "/_dist_/lib/lit-html.js";
import { GenreService } from "/_dist_/genre/GenreService.js";
import { SerieView } from "/_dist_/serie/SerieView.js";
import { BreadCrumbs } from "/_dist_/BreadCrumbs.js";

export class GenreView extends HTMLElement {
    constructor() {
        super();
        this.data = [];
        this.service = new GenreService();
        this.serieView = new SerieView();
    }

    connectedCallback() {
        addEventListener(this.service.getCustomEventName(), e => this.update(e));
        console.log("GenreView.connectedCallback");
        this.render();
    }

    update(event) {
        console.log("GenreView.update: Received event ", event);

        const e = new CustomEvent(BreadCrumbs.eventName(), [{url: "/", title: "Home"}]);
        this.dispatchEvent(e);
        console.log("GenreView.update: Dispatched event ", e);

        this.data = this.service.getData(event.detail.key);

        render(this.view(), this);
    }

    render() {
        var id = this.location.params.id;
        console.log("GenreView.render: running fetchData(" + id + ")");
        this.service.setId(id);
        this.service.fetchData();
    }

    view() {
        const items = [];
        this.data.forEach(item => {
            items.push("serie" ===  item.type ? this.renderSerieItem(item) : this.serieView.renderMovieItem(item));
        });

        return html`
            <div class="columns is-multiline is-mobile">
                ${items}
            </div>`;
    }

    renderSerieItem(item) {
        return html`
        <div class="column is-one-quarter-tablet is-half-mobile">
            <div class="card">
                <div class="card-image">
                    <figure class="image is-256x256">
                        <a href="/${item.type}/${item.id}">
                            <img class="is-rounded" src="${item.thumbnail}" alt="${item.title}"></img>
                        </a>
                    </figure>
                </div>
                <footer class="card-footer">
                    <a class="card-footer-item" href="/${item.type}/${item.id}">
                        ${item.title}
                    </a>
                </footer>
            </div>
        </div>`;
    }
}

customElements.define("genre-view", GenreView);
