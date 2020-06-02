import { RestClientService } from "../RestClientService.js";

export class GenreTableService extends RestClientService {
    constructor() {
        super();
    }

    getCustomEventName() {
        return "GenreTableService";
    }

    getServicePath() {
        return "/genres/count";
    }

    getItemsUrl(page, pageSize) {
        return this.getBaseUri() + "/genres/items?page=" + page + "&pageSize=" + pageSize;
    }
}
