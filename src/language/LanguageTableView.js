import { html, render } from "/_dist_/lib/lit-html.js";
import { LanguageTableService } from "/_dist_/language/LanguageTableService.js";

export class LanguageTableView extends HTMLElement {
    constructor() {
        super();
        this.data = [];
        this.service = new LanguageTableService();
    }

    connectedCallback() {
        addEventListener(this.service.getCustomEventName(), e => this.update(e));
        console.log("LanguageTableView.connectedCallback");
        this.render();
    }

    update(event) {
        console.log("LanguageTableView.update: Received event ", event);

        this.data = this.service.getData(event.detail.key);

        render(this.view(), this);

        let datatable = new DataTable("#languageTable", {
            columns: [{name: 'ID', id: "id", width: 128}, {name: 'Name', id: "name", width: 256}, {name: 'Order', id: "displayOrder", width: 128}],
            data: this.data.map(o => [o.id, o.name, o.displayOrder])
        });
    }

    render() {
        console.log("LanguageTableView.render: running fetchData()");
        this.service.fetchData();
    }

    view() {
        return html`
            <div id="languageTable">
            </div>`;
    }
}

customElements.define("language-table-view", LanguageTableView);
