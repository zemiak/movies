import { html, render } from "./lib/lit-html.js";

export class TemplateService {
    constructor(url) {
        this.baseUrl = url;
    }

    renderItems(data, element) {
        var plainHtml = "";

        plainHtml = plainHtml + this.header() + "\n";

        this.folders = false;
        data.forEach(item => {plainHtml = plainHtml + this.element(item)});

        plainHtml = plainHtml + this.footer() + "\n";

        if (! this.folders) {
            this.dispatchGalleryEvent();
        }

        // const template = html(plainHtml);
        // console.log(template);
        // render(template, element);

        element.innerHTML = plainHtml;
    }

    header() {


        return `
    <bread-crumbs></bread-crumbs>
    <ul class="auto-grid">
        `;
    }

    footer() {
        return `    </ul>
`;
    }

    element(item) {
        if ("genre" === item.type || "serie" === item.type) {
            return this.renderFolder(item);
        }

        if ("video" === item.type) {
            return this.renderVideo(item);
        }

        return this.renderPicture(item);
    }

    renderPicture(item) {
        const detailUrl = item.url;
        const thumbnailUrl = item.thumbnail;

        return `
        <li>
            <a href="${detailUrl}" data-type="image" class="glightbox" data-gallery="gallery1">
                <img src="${thumbnailUrl}" class="image-box" alt="${item.title}">
            </a>
        </li>
`;
    }

    renderVideo(item) {
        const detailUrl = item.url;
        const thumbnailUrl = item.thumbnail;

        return `
        <li>
            <video controls class="video-box">
                <source src="${detailUrl}" type="video/mp4"">
                <a href="${detailUrl}"><img src="${thumbnailUrl}" class="video-box" /></a>
            </video>
        </li>
`;
    }

    renderFolder(item) {
        const thumbnailUrl = item.thumbnail;

        this.folders = true;

        return `
        <li>
            <a href="#${item.path}">
                <img src="${thumbnailUrl}" class="folder-box" alt="${item.title}">
            </a>
            <span>${item.title}</span>
        </li>
`;
    }

    dispatchGalleryEvent() {
        const folderDataEvent = new CustomEvent(TemplateService.eventName(), {detail: {}, bubbles: true});
        dispatchEvent(folderDataEvent);
        console.log("TemplateService.dispatchGalleryEvent: fired");
    }

    static eventName() {
        return "render-lightbox-event";
    }
}
