import { RestClientService } from "../RestClientService.js";

export class MovieTableService extends RestClientService {
    constructor() {
        super();
    }

    getCustomEventName() {
        return "MovieTableService";
    }

    getServicePath() {
        return "/movies/all";
    }

    getCacheKeyName() {
        return "_" + this.getCustomEventName();
    }

    setId(id) {
        this.id = id;
    }
}
