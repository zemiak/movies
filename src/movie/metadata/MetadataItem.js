import { html, render } from "lit-html";

export class MetadataItem extends HTMLElement {
    connectedCallback() {
        this.render();
    }

    render() {
        let imageUrl = this.getAttribute("data-image-url");
        let description = this.getAttribute("data-description");
        let title = this.getAttribute("data-title");

        let template = html `<a href="javascript:void(0)" onclick="javascript:window._Metadata_detailClick(this, '${title}')" class="panel-block" style="height: 50px;">
        <span class="panel-icon">
          <figure class="image is-16x16"><img src="${imageUrl}"></figure>
        </span>
        <div style="padding-left: 1em;">${description}</div>
      </a>`;

        render(template, this);
    }
}

customElements.define("metadata-item", MetadataItem);
