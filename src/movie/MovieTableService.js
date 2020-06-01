import { RestClientService } from "../RestClientService.js";

export class MovieTableService extends RestClientService {
    constructor() {
        super();
    }

    getCustomEventName() {
        return "MovieTableService";
    }

    getServicePath() {
        return "/movies/count";
    }

    getItemsUrl(page, pageSize) {
        return this.getBaseUri() + "/movies/items?page=" + page + "&pageSize=" + pageSize;
    }
}
