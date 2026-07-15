using System.Text.Json.Serialization;

namespace Dragonfly;

[JsonSourceGenerationOptions(PropertyNameCaseInsensitive = false)]
[JsonSerializable(typeof(Login.IdentityData))]
[JsonSerializable(typeof(Login.ClientData))]
internal sealed partial class LoginJsonContext : JsonSerializerContext;
