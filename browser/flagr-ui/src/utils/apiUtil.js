import axios from "axios"
import constants from '@/constants'
import store from '../store';  // Adjust the path to your store
import router from "../router";

let axiosInstance ;
let flagrAxiosInstance;

const API_URLS = constants.API_URLS
export const setupAxiosInstance = (baseUrl) => {
	axiosInstance = axios.create({
		baseURL: baseUrl,
		timeout: 300000,
		headers: headerParams(),
	})
	//setupAxiosInstanceInterceptors()
	return axiosInstance
}

export const setupflagrAxiosInstance = () => {
	flagrAxiosInstance = axios.create({
		baseURL: window.location.origin+'/api/v1',
		timeout: 300000,
		headers: headerParams(),
	})
	//setupAxiosInstanceInterceptors()
	return flagrAxiosInstance
}

const deviceType = ["web", "mweb", "ios-webview", "android-webview"]

export const detectDeviceType = () => {
	if (typeof window !== "undefined") {
		const details = navigator.userAgent
		const regexp = /android|iphone|kindle|ipad/i
		const isMobileDevice = regexp.test(details)
		if (isMobileDevice) {
			return deviceType[1]
		} else {
			return deviceType[0]
		}
	}
}

const headerParams = (default_headers = {}) => {
	const headers = default_headers
	headers["Content-Type"] = "application/json"
	headers["x-client-type"] = detectDeviceType() ?? "unknown"
	headers["accept"] = "application/json"
	if (typeof window !== "undefined") {
		let uuid = localStorage.getItem("uuid")
		if (!uuid) {
			uuid = generateUUID()
			localStorage?.setItem("uuid", uuid)
		}
		headers["x-device-id"] = uuid
	}
	return headers
}


export const generateUUID = () => {
	let d = new Date().getTime()
	let d2 = (typeof performance !== "undefined" && performance.now && performance.now() * 1000) || 0 //Time in microseconds since page-load or 0 if unsupported
	return "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".replace(/[xy]/g, function (c) {
		let r = Math.random() * 16 //random number between 0 and 16
		if (d > 0) {
			//Use timestamp until depleted
			r = (d + r) % 16 | 0
			d = Math.floor(d / 16)
		} else {
			//Use microseconds since page-load if supported
			r = (d2 + r) % 16 | 0
			d2 = Math.floor(d2 / 16)
		}
		return (c === "x" ? r : (r & 0x3) | 0x8).toString(16)
	})
}

export const getAxiosInstance = () => {
	if (!axiosInstance) {
		throw new Error("Axios instance not initialized")
	}
	const token = localStorage.getItem('tokens');

	// Try to parse the token and catch any potential errors
	if(token){
		try {
			const authToken = token ? JSON.parse(token) : null;
			if (authToken && authToken["x-access-token"]) {
				axiosInstance.defaults.headers["Authorization"] = "Bearer " + authToken["x-access-token"];
			}
		} catch (error) {
			console.error("Failed to parse token from localStorage", error);
		}
	} else {
		delete axiosInstance.defaults.headers["Authorization"]
	}

	// Set the Authorization header if the token exist

	return axiosInstance;
}


export const getAxiosFlagrInstance = () => {
	if (!flagrAxiosInstance) {
		throw new Error("Axios instance not initialized")
	}
	const token = localStorage.getItem('tokens');

	// Try to parse the token and catch any potential errors
	if(token){
		try {
			const authToken = token ? JSON.parse(token) : null;
			if (authToken && authToken["x-access-token"]) {
				flagrAxiosInstance.defaults.headers["Authorization"] = "Bearer " + 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJhVVNzVzhHSTAzZHlRMEFJRlZuOTIiLCJkX3R5cGUiOiJ3ZWIiLCJkaWQiOiJmOTMxNTU1MS01NTZlLTQ5ZmUtOTAyZS05YmUxYzE2MzYzYzIiLCJlX2lkIjoiNTQzNjIyODYzIiwiZXhwIjoxNzI3NDIyOTg0LCJpYXQiOiIyMDI0LTA5LTI2VDA3OjQzOjA0LjM4NDE4ODMzNVoiLCJpc3MiOiJhdXRoZW50aWNhdGlvbi5hbGxlbi1zYW5kYm94IiwiaXN1IjoiIiwicHQiOiJJTlRFUk5BTF9VU0VSIiwic2lkIjoiYzUzNjFhM2ItOWZhMy00OGZlLTkyOWUtMDBhY2M3YjUzYjMxIiwidGlkIjoiYVVTc1c4R0kwM2R5UTBBSUZWbjkyIiwidHlwZSI6ImFjY2VzcyIsInVpZCI6ImhaSkx1andPVmZIM3cybEFQVE1YTCJ9._37wBjStkn1GxlEcRrepGT03j6DmNp8jKBmBB_HAFDE';
			}
		} catch (error) {
			console.error("Failed to parse token from localStorage", error);
		}
	} else {
		delete flagrAxiosInstance.defaults.headers["Authorization"]
	}

	// Set the Authorization header if the token exist

	return flagrAxiosInstance;
}


export const getGoogleRedirectionLink = async (postData) => {
    const response = await getAxiosInstance().get(
		`${`api/v1/auth/authenticate/googleAuthRedirectLink`}?uiRedirectUrl=${postData.uiRedirectUrl}`,
	)
	const redirectUrlResponse = response?.data?.data
	return redirectUrlResponse
}


export const googleAuthenticate = async (postData) => {	
	const reqPayload = JSON.stringify(postData)
	try {
		const response = await getAxiosInstance().post(`api/v1/auth/authenticate/google`, reqPayload)
		if (typeof window !== "undefined") {
			const token = {}
			token["x-access-token"] = response.headers["x-access-token"]
			token["x-refresh-token"] = response.headers["x-refresh-token"]
			localStorage.setItem("tokens", JSON.stringify(token))
		}
		return response?.data?.data
	} catch (error) {
		console.error(error)
		if (typeof window !== "undefined") {
			localStorage.removeItem("tokens")
			localStorage.removeItem("chat-font-size")
		}
		throw error
	}
}

export function isAuthenticated() {
	return !!localStorage.getItem('tokens'); // Assuming token is saved in localStorage
}

export async function logout() {
	
	try {
		await getAxiosInstance().post(`api/v1/auth/logout`)
	} catch (error) {
		alert('Couldnot logout');
	} finally {
		store.dispatch("reset")
		if (typeof window !== "undefined") {
			localStorage.removeItem("tokens")
			console.log('logging out in logout')
			router.replace({ name: 'Login' })
		}
	}
}

export async function getUserDetails() {
	const response = await getAxiosInstance().get(`${API_URLS.USER_DETAILS}`)
	console.log(response.data.data)
	return response.data.data
}