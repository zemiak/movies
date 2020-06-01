import { RestClientService } from "../RestClientService.js";

export class SerieTableService extends RestClientService {
    constructor() {
        super();
    }

    getCustomEventName() {
        return "SerieTableService";
    }

    getServicePath() {
        return "/series/count";
    }

    getItemsUrl(page, pageSize) {
        return this.getBaseUri() + "/movies/items?page=" + page + "&pageSize=" + pageSize;
    }
}
