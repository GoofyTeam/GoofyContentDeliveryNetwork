-- Configuration des requêtes
request = function()
    -- Ajouter des headers aléatoires pour éviter le cache
    local path = "/test/cache/static/" .. math.random(1, 1000)
    local headers = {}
    return wrk.format("GET", path, headers)
end

-- Fonction appelée quand un thread démarre
init = function(args)
    math.randomseed(os.time())
end

-- Fonction appelée quand une réponse est reçue
response = function(status, headers, body)
    if status ~= 200 then
        wrk.thread:stop()
    end
end
