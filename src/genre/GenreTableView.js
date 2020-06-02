import { html, render } from "lit-html";
import { GenreTableService } from "/_dist_/genre/GenreTableService.js";
import '@vaadin/vaadin-grid/vaadin-grid';

export class GenreTableView extends HTMLElement {
    constructor() {
        super();
        this.data = [];
        this.service = new GenreTableService();
    }

    connectedCallback() {
        addEventListener(this.service.getCustomEventName(), e => this.update(e));
        console.log("GenreTableView.connectedCallback");
        this.render();
    }

    update(event) {
        console.log("GenreTableView.update: Received event ", event);

        this.data = this.service.getData(event.detail.key);

        console.log("GenreTableView.update: Data ", this.data);

        render(this.view(), this);

        var grid = document.querySelector('#genreTable');
        grid.size = this.data.count;

        var that = this;
        grid.dataProvider = function(params, callback) {
            var xhr = new XMLHttpRequest();
            xhr.onload = function() {
                console.log("GenreTableView.update: XHR Data ", xhr.responseText);
                callback(JSON.parse(xhr.responseText).result);
            };

            var index = params.page * params.pageSize;
            xhr.open('GET', that.service.getItemsUrl(params.page, params.pageSize), true);
            xhr.send();
        };
    }

    render() {
        console.log("GenreTableView.render: running fetchData()");
        this.service.fetchData();
    }

    view() {
        return html`
    <vaadin-grid id="genreTable">
        <vaadin-grid-column path="displayOrder" header="Order" width="7em" flex-grow="0"></vaadin-grid-column>
        <vaadin-grid-column path="name" header="Name"></vaadin-grid-column>
        <vaadin-grid-column path="protectedGenre" header="Protected"></vaadin-grid-column>
        <vaadin-grid-column path="id" header="ID" width="7em" flex-grow="0"></vaadin-grid-column>
    </vaadin-grid>`;
    }
}

customElements.define("genre-table-view", GenreTableView);
