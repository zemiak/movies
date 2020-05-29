import { html, render } from "/_dist_/lib/lit-html.js";
import { SerieTableService } from "/_dist_/serie/SerieTableService.js";

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

        render(this.view(), this);

        let datatable = new DataTable("#serieTable", {
            columns: [{name: 'ID', id: "id", width: 128}, {name: 'Name', id: "name", width: 256}, {name: 'Order', id: "displayOrder", width: 128}],
            data: this.data.map(o => [o.id, o.name, o.displayOrder])
        });
    }

    render() {
        console.log("SerieTableView.render: running fetchData()");
        this.service.fetchData();
    }

    view() {
        return html`
            <div id="serieTable">
            </div>`;
    }
}

customElements.define("serie-table-view", SerieTableView);
