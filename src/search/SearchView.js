import { html, render } from "/_dist_/lib/lit-html.js";
import { SearchService } from "/_dist_/search/SearchService.js";
import { SerieView } from "/_dist_/serie/SerieView.js";
import { GenreView } from "/_dist_/genre/GenreView.js"
import { BreadCrumbs } from "/_dist_/BreadCrumbs.js";

export class SearchView extends HTMLElement {
    constructor() {
        super();
        this.data = [];
        this.service = new SearchService();
        this.serieView = new SerieView();
        this.genreView = new GenreView();
    }

    connectedCallback() {
        addEventListener(this.service.getCustomEventName(), e => this.update(e));
        console.log("SearchView.connectedCallback");
        this.render();
    }

    update(event) {
        console.log("SearchView.update: Received event ", event);

        const e = new CustomEvent(BreadCrumbs.eventName(), [{url: "/", title: "Home"}]);
        this.dispatchEvent(e);
        console.log("SearchView.update: Dispatched event ", e);

        this.data = this.service.getData(event.detail.key);

        render(this.view(), this);
    }

    render() {
        var query = this.location.params.query;
        console.log("SearchView.render: running fetchData(" + query + ")");
        this.service.setQuery(query);
        this.service.fetchData();
    }

    view() {
        const items = [];
        this.data.forEach(item => {
            items.push("serie" ===  item.type ? this.genreView.renderSerieItem(item) : this.serieView.renderMovieItem(item));
        });

        return html`
            <div class="columns is-multiline is-mobile">
                ${items}
            </div>`;
    }
}

customElements.define("search-view", SearchView);
