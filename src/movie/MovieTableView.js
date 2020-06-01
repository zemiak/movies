import { html, render } from "lit-html";
import { MovieTableService } from "/_dist_/movie/MovieTableService.js";

export class MovieTableView extends HTMLElement {
    constructor() {
        super();
        this.data = [];
        this.service = new MovieTableService();
    }

    connectedCallback() {
        addEventListener(this.service.getCustomEventName(), e => this.update(e));
        console.log("MovieTableView.connectedCallback");
        this.render();
    }

    update(event) {
        console.log("MovieTableView.update: Received event ", event);

        this.data = this.service.getData(event.detail.key);

        render(this.view(), this);

        let datatable = new DataTable("#movieTable", {
            columns: [{name: 'ID', id: "id", width: 128}, {name: 'Name', id: "name", width: 256}, {name: 'Order', id: "displayOrder", width: 128}],
            data: this.data.map(o => [o.id, o.name, o.displayOrder])
        });
    }

    render() {
        console.log("MovieTableView.render: running fetchData()");
        this.service.fetchData();
    }

    view() {
        return html`
            <div id="movieTable">
            </div>`;
    }
}

customElements.define("movie-table-view", MovieTableView);
