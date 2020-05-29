export class BreadCrumbs extends HTMLElement {
    constructor() {
        super();
        this.items = [{url: "/", title: "Home"}];
    }

    connectedCallback() {
        addEventListener(BreadCrumbs.eventName(), e => this.update(e));
        this.render();
    }

    update(event) {
        console.log("BreadCrumbs.update: received event ", event);
        this.items = event.data;
        this.render();
    }

    render() {
        var html = '<div class="level-item"><nav class="breadcrumb" aria-label="breadcrumbs"><ul>';
        this.items.forEach((item) => {
            html = html + this.renderItem(item);
        })
        html += '</ul></nav></div>';

        this.innerHTML = html;
    }

    renderItem(item) {
        return `<li><a href="${item.url}">${item.title}</a></li>`;
    }

    static eventName() {
        return "BreadCrumbs";
    }
}

customElements.define("bread-crumbs", BreadCrumbs);
