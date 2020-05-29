import { RestClientService } from "../RestClientService.js";

export class MovieService extends RestClientService {
    constructor() {
        super();
    }

    getCustomEventName() {
        return "MovieService";
    }

    getServicePath() {
        return "/movies/" + this.id;
    }

    getCacheKeyName() {
        return "_" + this.getCustomEventName() + "_" + this.id;
    }

    setId(id) {
        this.id = id;
    }
}
