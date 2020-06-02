import { html, render } from "lit-html";
import { SerieTableService } from "/_dist_/serie/SerieTableService.js";
import '@vaadin/vaadin-grid/vaadin-grid';

export class SerieTableView extends HTMLElement {
    constructor() {
        super();
        this.data = [];
        this.service = new SerieTableService();
    }

    connectedCallback() {
        addEventListener(this.service.getCustomEventName(), e => this.update(e));
        console.log("SerieTableView.connectedCallback");
        this.render();
    }

    update(event) {
        console.log("SerieTableView.update: Received event ", event);

        this.data = this.service.getData(event.detail.key);

        console.log("SerieTableView.update: Data ", this.data);

        render(this.view(), this);

        var grid = document.querySelector('#genreTable');
        grid.size = this.data.count;

        var that = this;
        grid.dataProvider = function(params, callback) {
            var xhr = new XMLHttpRequest();
            xhr.onload = function() {
                console.log("SerieTableView.update: XHR Data ", xhr.responseText);
                callback(JSON.parse(xhr.responseText).result);
            };

            var index = params.page * params.pageSize;
            xhr.open('GET', that.service.getItemsUrl(params.page, params.pageSize), true);
            xhr.send();
        };
    }

    render() {
        console.log("SerieTableView.render: running fetchData()");
        this.service.fetchData();
    }

    view() {
        return html`
    <vaadin-grid id="genreTable">
        <vaadin-grid-column path="displayOrder" header="Order" width="7em" flex-grow="0"></vaadin-grid-column>
        <vaadin-grid-column path="name" header="Name"></vaadin-grid-column>
        <vaadin-grid-column path="genre" header="Genre"></vaadin-grid-column>
        <vaadin-grid-column path="tvShow" header="TV Show"></vaadin-grid-column>
        <vaadin-grid-column path="id" header="ID" width="7em" flex-grow="0"></vaadin-grid-column>
    </vaadin-grid>`;
    }
}

customElements.define("serie-table-view", SerieTableView);
