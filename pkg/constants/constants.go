package constants

// Database constants
const (
	TableName           = "urls"
	IndexNameShortCode  = "idx_urls_short_code"
	IndexNameUserID     = "idx_urls_user_id"
	DatabaseInitialized = "Database initialized successfully"
	InitializingDB      = "Initializing database..."
	TableCreated        = "URLs table created successfully"
	TableExists         = "URLs table already exists, checking schema..."
	TableRecreated      = "Table recreated successfully"
	TableSchemaValid    = "Table schema is valid"
	TableSchemaInvalid  = "Table schema is invalid, recreating table..."
	IndexCreating       = "Creating index on short_code..."
	IndexCreated        = "Creating 'urls' table..."
	CleanupScheduled    = "URL cleanup scheduler started (runs every 10 minutes)"
	CleanupError        = "Error during scheduled cleanup:"
	CleanupCompleted    = "Cleaned up %d old URLs (older than 1 hour)"
	DashboardNote       = "NOTE: To see data in PocketBase dashboard:"
	DashboardStep1      = "1. Go to http://localhost:8091/_/"
	DashboardStep2      = "2. Create a new collection named 'urls'"
	DashboardStep3      = "3. Add fields: short_code (text), original_url (text), clicks (number)"
	DashboardStep4      = "4. The data will then be visible in the dashboard"
)

// Query constants
const (
	ShortCodeFilter = "short_code = {:shortCode}"
	UpdatedFilter   = "updated < {:cutoffTime}"
	CreatedSortDesc = "-created"
	NoFilter        = ""
	NoSort          = ""
	RecordBatchSize = 500
	LargeBatchSize  = 50000
	NoOffset        = 0
)

// Parameter names for database queries
const (
	ParamShortCode   = "shortCode"
	ParamOriginalURL = "originalURL"
	ParamClicks      = "clicks"
	ParamCreatedAt   = "createdAt"
	ParamUpdatedAt   = "updatedAt"
	ParamCutoffTime  = "cutoffTime"
	ParamUserID      = "userId"
)

// Column names
const (
	ColumnID          = "id"
	ColumnShortCode   = "short_code"
	ColumnOriginalURL = "original_url"
	ColumnClicks      = "clicks"
	ColumnCreated     = "created"
	ColumnUpdated     = "updated"
	ColumnUserID      = "user_id"
)

// URL and protocol constants
const (
	DefaultBaseURL  = "http://localhost:8090"
	APIBasePath     = "/api"
	WebSocketPath   = "/ws"
	HTTPSProtocol   = "https://"
	HTTPProtocol    = "http://"
	WSSProtocol     = "wss://"
	WSProtocol      = "ws://"
	DefaultProtocol = "https://"
	EnvBaseURL      = "BASE_URL"
)

// Time constants
const (
	TimeFormat         = "2006-01-02 15:04:05"
	CleanupInterval    = 10 // minutes
	URLExpirationHours = 1  // hour
)

// Base62 encoding constants
const (
	Base62Charset   = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	ShortCodeLength = 6
)

// WebSocket constants
const (
	WSMessageType   = "type"
	WSStatsUpdate   = "stats_update"
	WSShortCode     = "short_code"
	WSClicks        = "clicks"
	WSTimestamp     = "timestamp"
	WSBroadcastFull = "Broadcast channel is full, skipping message"
	WSUpgradeError  = "WebSocket upgrade error:"
	WSReadError     = "WebSocket read error:"
	WSWriteError    = "WebSocket write error:"
	WSConnected     = "Client connected:"
	WSDisconnected  = "Client disconnected:"
	WSMarshalError  = "Error marshaling stats update:"
	PBIDPrefix      = "r"
)

// HTTP status messages
const (
	HelloMessage        = "Hello from PocketBase!"
	SuccessStatus       = "success"
	InvalidJSON         = "Invalid JSON data"
	URLRequired         = "URL is required"
	InvalidURLFormat    = "Invalid URL format"
	HTTPSOnly           = "Only HTTP and HTTPS protocols are supported"
	URLNotFound         = "URL not found"
	ShortCodeRequired   = "Short code not provided"
	InvalidStoredURL    = "Invalid URL stored"
	StoreURLError       = "Failed to store URL"
	FetchURLError       = "Failed to fetch recent URLs"
	IncrementError      = "Failed to increment click count"
	DatabaseError       = "Failed to get updated data"
	CleanupErrorMsg     = "Failed to cleanup old URLs"
	CleanupCompletedMsg = "Cleanup completed"
	ManualCleanup       = "Manual cleanup triggered"
	TestMessage         = "Click count test successful"
)

// JSON field names
const (
	JSONMessage       = "message"
	JSONStatus        = "status"
	JSONError         = "error"
	JSONOriginalURL   = "original_url"
	JSONShortCode     = "short_code"
	JSONShortURL      = "short_url"
	JSONClicks        = "clicks"
	JSONCreated       = "created"
	JSONRowsDeleted   = "rows_deleted"
	JSONCutoffTime    = "cutoff_time"
	JSONCleanupReason = "cleanup_reason"
	JSONTest          = "test"
	JSONType          = "type"
	JSONData          = "data"
	JSONTimestamp     = "timestamp"
	JSONPrevClicks    = "previous_clicks"
	JSONCurrClicks    = "current_clicks"
	JSONClicksInc     = "clicks_incremented"
	JSONUser          = "user"
)

// Response field names (for API responses)
const (
	ResponseID          = "id"
	ResponseShortCode   = "short_code"
	ResponseOriginalURL = "original_url"
	ResponseClicks      = "clicks"
	ResponseCreated     = "created"
	ResponseShortURL    = "short_url"
)

// Cache control headers
const (
	CacheControl  = "Cache-Control"
	Pragma        = "Pragma"
	Expires       = "Expires"
	LastModified  = "Last-Modified"
	XTimestamp    = "X-Timestamp"
	NoCache       = "no-cache, no-store, must-revalidate"
	NoCacheMaxAge = "no-cache, no-store, must-revalidate, max-age=0"
	NoStore       = "no-cache, no-store, must-revalidate"
)

// Additional JSON field names
const (
	JSONErrorKey = "error"
	JSONUrls     = "urls"
)

// Additional error messages
const (
	ErrorShortCodeNotProvided = "Short code not provided"
	ErrorURLNotFound          = "URL not found"
	ErrorInvalidURLStored     = "Invalid URL stored"
	ErrorUnauthorized         = "Unauthorized access"
	ErrorInvalidCredentials   = "Invalid email or password"
	ErrorUserExists           = "User already exists"
	ErrorEmailRequired        = "Email is required"
	ErrorPasswordRequired     = "Password is required"
	ErrorPasswordTooShort     = "Password must be at least 6 characters"
	ErrorInvalidEmail         = "Invalid email format"
	ErrorTokenExpired         = "Authentication token expired"
	ErrorTokenInvalid         = "Invalid authentication token"
	ErrorSessionRequired      = "Authentication required"
)

// HTTP header values
const (
	HeaderValueZero       = "0"
	HTTPTimeFormat        = "Mon, 02 Jan 2006 15:04:05 GMT"
	WebSocketTextMessage  = 1
	WebSocketCloseMessage = 8
)

// URL scheme constants
const (
	HTTPScheme  = "http"
	HTTPSScheme = "https"
)

// JSON struct field names
const (
	JSONFieldURL      = "url"
	JSONFieldEmail    = "email"
	JSONFieldPassword = "password"
	JSONFieldToken    = "token"
	JSONFieldUser     = "user"
	JSONFieldUserID   = "user_id"
)

// Path parameter names
const (
	PathParamShortCode = "shortCode"
)
