import { RestClientService } from "../RestClientService.js";

export class SearchService extends RestClientService {
    constructor() {
        super();
    }

    getCustomEventName() {
        return "GenreService";
    }

    getServicePath() {
        return "/ui/search?q=" + this.query;
    }

    getCacheKeyName() {
        return null; // no caching
    }

    setQuery(query) {
        this.query = encodeURIComponent(query);
    }
}
