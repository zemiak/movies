import { RestClientService } from "../RestClientService.js";

export class MovieAddService extends RestClientService {
    getCustomEventName() {
        return "MovieAddService";
    }

    getServicePath() {
        return "/movies/detail/forNew";
    }
}
