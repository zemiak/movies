import { RestClientService } from "../RestClientService.js";

export class GenreService extends RestClientService {
    constructor() {
        super();
    }

    getCustomEventName() {
        return "GenreService";
    }

    getServicePath() {
        return "/genres/browse?id=" + this.id;
    }

    getCacheKeyName() {
        return "_" + this.getCustomEventName() + "_" + this.id;
    }

    setId(id) {
        this.id = id;
    }
}
