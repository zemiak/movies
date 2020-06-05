import { html, render } from "lit-html";
import { RenderLanguageDetail } from "/_dist_/language/RenderLanguageDetail.js";
import { LanguageDetailService } from "/_dist_/language/LanguageDetailService.js";

export class LanguageDetailView extends HTMLElement {
    constructor() {
        super();
        this.data = {};
        this.service = new LanguageDetailService();
        this.renderer = new RenderLanguageDetail();
    }

    connectedCallback() {
        addEventListener(this.service.getCustomEventName(), e => this.update(e));
        console.log("LanguageDetailView.connectedCallback");
        this.render();
    }

    update(event) {
        console.log("LanguageDetailView.update: Received event ", event);

        const e = new CustomEvent(BreadCrumbs.eventName(), [{url: "/", title: "Home"}]);
        this.dispatchEvent(e);
        console.log("LanguageDetailView.update: Dispatched event ", e);

        this.data = this.service.getData(event.detail.key);

        render(this.view(), this);
    }

    render() {
        var id = this.location.params.id;
        console.log("LanguageDetailView.render: running fetchData(" + id + ")");
        this.service.setId(id);
        this.service.fetchData();
    }

    view() {
        var entity = this.data;
        var form = this.renderer.render(entity, true, false);
        return html`${form}`;
    }
}

customElements.define("language-detail-view", LanguageDetailView);
