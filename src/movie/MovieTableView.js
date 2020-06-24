import { html, render } from "lit-html";
import { MovieTableService } from "/_dist_/movie/MovieTableService.js";
import '@vaadin/vaadin-grid/vaadin-grid';

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

        console.log("MovieTableView.update: Data ", this.data);

        render(this.view(), this);

        var grid = document.querySelector('#movieTable');
        grid.size = this.data.count;

        var that = this;
        document.querySelector('#editColumn').renderer = (root, grid, rowData) => {
            root.innerHTML = `<button class="button is-info is-small" onclick="window.location='/admin/movie/${rowData.item.id}'">
                <span class="icon is-small">
                    <i class="fas fa-edit" aria-hidden="true"></i>
                </span>
                <span>Edit</span>
            </button>`;
        };

        grid.dataProvider = function(params, callback) {
            var xhr = new XMLHttpRequest();
            xhr.onload = function() {
                console.log("MovieTableView.update: XHR Data ", xhr.responseText);
                callback(JSON.parse(xhr.responseText).result);
            };

            var index = params.page * params.pageSize;
            xhr.open('GET', that.service.getItemsUrl(params.page, params.pageSize), true);
            xhr.send();
        };
    }

    render() {
        console.log("MovieTableView.render: running fetchData()");
        this.service.fetchData();
    }

    view() {
        return html`
    <vaadin-grid id="movieTable">
        <vaadin-grid-column path="name" header="Name"></vaadin-grid-column>
        <vaadin-grid-column path="serie" header="Serie"></vaadin-grid-column>
        <vaadin-grid-column path="displayOrder" header="Serie Order" width="7em" flex-grow="0"></vaadin-grid-column>
        <vaadin-grid-column path="id" header="ID" width="7em" flex-grow="0"></vaadin-grid-column>
        <vaadin-grid-column id="editColumn" header="" width="6em" flex-grow="0"></vaadin-grid-column>
    </vaadin-grid><p>&nbsp;</p>
    <button class="button is-info" onclick="window.location='/admin/movie/add'">
        <span class="icon is-small">
            <i class="fas fa-plus" aria-hidden="true"></i>
        </span>
        <span>Add</span>
    </button>`;
    }
}

customElements.define("movie-table-view", MovieTableView);
