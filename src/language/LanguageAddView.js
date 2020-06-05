import { html, render } from "lit-html";
import { RenderLanguageDetail } from "/_dist_/language/RenderLanguageDetail.js";

export class LanguageAddView extends HTMLElement {
    connectedCallback() {
        this.render();
        this.renderer = new RenderLanguageDetail();
    }

    render() {
        render(this.view(), this);
    }

    view() {
        var entity = {id: null, code: "", name: "", displayOrder: 0};
        var form = this.renderer.render(entity, true, false);
        return html`${form}`;
    }
}

customElements.define("language-add-view", LanguageAddView);
