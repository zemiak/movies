import { html, render } from "lit-html";

export class MetadataListDialog extends HTMLElement {
    constructor() {
        this.toggle = this.toggle.bind(this);
        this.on = this.on.bind(this);
        this.off = this.off.bind(this);
    }

    connectedCallback() {
        this.render();
    }

    render() {
        let title = this.getAttribute("data-title");

        let template = html `<div class="modal" id="${title}-list-dialog">
        <div class="modal-background"></div>
        <div class="modal-card">
        <header class="modal-card-head">
            <p class="modal-card-title">${title} Metadata</p>
            <button class="delete" aria-label="close" type="button" @click="${this.off}"></button>
        </header>
        <section class="modal-card-body">
            <div id="${title}-list"></div>
        </section>
        </div>
    </div>`;

        render(template, this);
    }

    toggle() {
        this.classList.toggle("is-active");
    }

    on() {
        if (!this.classList.contains("is-active")) {
            this.classList.add("is-active");
        }
    }

    off() {
        if (this.classList.contains("is-active")) {
            this.classList.remove("is-active");
        }
    }
}

customElements.define("metadata-list-dialog", MetadataListDialog);
