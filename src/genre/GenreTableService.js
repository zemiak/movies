import { RestClientService } from "../RestClientService.js";

export class GenreTableService extends RestClientService {
    constructor() {
        super();
    }

    getCustomEventName() {
        return "GenreTableService";
    }

    getServicePath() {
        return "/genres/all";
    }

    getCacheKeyName() {
        return "_" + this.getCustomEventName();
    }

    setId(id) {
        this.id = id;
    }
}
