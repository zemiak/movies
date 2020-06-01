import { RestClientService } from "../RestClientService.js";

export class LanguageTableService extends RestClientService {
    constructor() {
        super();
    }

    getCustomEventName() {
        return "LanguageTableService";
    }

    getServicePath() {
        return "/languages/count";
    }

    getItemsUrl(page, pageSize) {
        return this.getBaseUri() + "/languages/items?page=" + page + "&pageSize=" + pageSize;
    }
}
