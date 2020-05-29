import { html, render } from "../lib/lit-html.js";

export class AboutView extends HTMLElement {
    connectedCallback() {
        this.render();
    }

    render() {
        render(this.view(), this);
    }

    view() {
        return html`<h1>About</h1>`;
    }
}

customElements.define("about-view", AboutView);
