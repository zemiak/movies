import { html, render } from "lit-html";

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
