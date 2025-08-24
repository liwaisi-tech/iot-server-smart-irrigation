from dependency_injector import containers, providers

from src.config.settings import Settings


class Container(containers.DeclarativeContainer):
    """Dependency injection container for the IoT Smart Irrigation System."""
    
    wiring_config = containers.WiringConfiguration(
        modules=[
            "src.app.factory",
        ]
    )
    
    # Configuration
    config = providers.Configuration()
    
    # Settings
    settings = providers.Singleton(Settings)