import { html, render } from "/_dist_/lib/lit-html.js";
import { SerieService } from "/_dist_/serie/SerieService.js";
import { BreadCrumbs } from "/_dist_/BreadCrumbs.js";

export class SerieView extends HTMLElement {
    constructor() {
        super();
        this.data = [];
        this.service = new SerieService();
    }

    connectedCallback() {
        addEventListener(this.service.getCustomEventName(), e => this.update(e));
        console.log("SerieView.connectedCallback");
        this.render();
    }

    update(event) {
        console.log("SerieView.update: Received event ", event);

        const e = new CustomEvent(BreadCrumbs.eventName(), [{url: "/", title: "Home"}]);
        this.dispatchEvent(e);
        console.log("SerieView.update: Dispatched event ", e);

        this.data = this.service.getData(event.detail.key);

        render(this.view(), this);
    }

    render() {
        var id = this.location.params.id;
        console.log("SerieView.render: running fetchData(" + id + ")");
        this.service.setId(id);
        this.service.fetchData();
    }

    view() {
        const items = [];
        this.data.forEach(item => {
            items.push(this.renderMovieItem(item));
        });

        return html`
            <div class="columns is-multiline is-mobile">
                ${items}
            </div>`;
    }

    renderMovieItem(item) {
        console.log("movie", item);
        return html`
        <div class="column is-one-quarter-tablet is-half-mobile">
            <div class="card">
                <div class="card-image">
                    <figure class="image is-3by4">
                        <a href="/${item.type}/${item.id}">
                            <img src="${item.thumbnail}" alt="${item.title}"></img>
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

customElements.define("serie-view", SerieView);
