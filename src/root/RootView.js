import { html, render } from "lit-html";
import { RootService } from "/_dist_/root/RootService.js";
import { BreadCrumbs } from "/_dist_/BreadCrumbs.js";

export class RootView extends HTMLElement {
    constructor() {
        super();
        this.data = [];
        this.service = new RootService();
    }

    connectedCallback() {
        console.log("RootView.connectedCallback");
        addEventListener(this.service.getCustomEventName(), e => this.update(e));
        this.service.fetchData();
    }

    update(event) {
        console.log("RootView.update: Received event ", event);

        const e = new CustomEvent(BreadCrumbs.eventName(), [{url: "/", title: "Home"}]);
        this.dispatchEvent(e);
        console.log("RootView.update: Dispatched event ", e);

        this.data = this.service.getData(event.detail.key);
        this.render();
    }

    render() {
        console.log("RootView.render: running lit-html render");
        render(this.view(), this);
    }

    view() {
        const items = [];
        this.data.forEach(item => {
            items.push(this.renderGenreItem(item));
        });

        return html`
            <div class="columns is-multiline is-mobile">
                ${items}
            </div>`;
    }

    renderGenreItem(item) {
        return html`
        <div class="column is-one-quarter-tablet is-half-mobile">
            <div class="card">
                <div class="card-image">
                    <figure class="image is-256x256">
                        <a href="/${item.type}/${item.id}">
                            <img class="is-rounded" src="${item.thumbnail}" alt="${item.title}">
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

customElements.define("root-view", RootView);
