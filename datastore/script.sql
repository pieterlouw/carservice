CREATE TABLE IF NOT EXISTS Bookings (
    ID INTEGER PRIMARY KEY,
	CustomerName TEXT NOT NULL,
    ContactNumber TEXT NOT NULL,
	CarRegistrationNumber TEXT NOT NULL,
    CarMake TEXT NOT NULL,
    CarModel TEXT NOT NULL,
    Odometer INTEGER NOT NULL,
    ServiceDate datetime NOT NULL,
	BookingDate datetime NOT NULL DEFAULT((STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')))
); 