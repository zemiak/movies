import { html, render } from "/_dist_/lib/lit-html.js";
import { GenreTableService } from "/_dist_/genre/GenreTableService.js";

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

        render(this.view(), this);

        let datatable = new DataTable("#genreTable", {
            columns: [{name: 'ID', id: "id", width: 128}, {name: 'Name', id: "name", width: 256}, {name: 'Order', id: "displayOrder", width: 128}],
            data: this.data.map(o => [o.id, o.name, o.displayOrder])
        });
    }

    render() {
        console.log("GenreTableView.render: running fetchData()");
        this.service.fetchData();
    }

    view() {
        return html`
            <div id="genreTable">
            </div>`;
    }
}

customElements.define("genre-table-view", GenreTableView);
