import { html, render } from "lit-html";

export class UnknownView extends HTMLElement {
    connectedCallback() {
        this.render();
    }

    render() {
        render(this.view(), this);
    }

    view() {
        return html`<h1>Unknown</h1>`;
    }
}

customElements.define("unknown-view", UnknownView);
